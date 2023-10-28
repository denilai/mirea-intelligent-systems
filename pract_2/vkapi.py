from typing import List, Any, Dict, Tuple, Optional, Set
import requests
import os
import json
import sys
import time
import urllib.parse


from utils import create_logger, parse_config

logger = create_logger("VkAPI")


"""Класс, инкапсулирующий работу с Vk APi (https://dev.vk.com/ru/reference)"""
class VkApiAgent:

    def __init__(self, endpoint:str, access_token:str):
        dir_path = os.path.dirname(os.path.realpath(__file__))
        self.api_errors = parse_config(os.path.join(dir_path, "api_errors.yaml"))

        self.api_endpoint = endpoint
        self.access_token = access_token

        self.session = requests.Session()
        self.session.hooks["response"].append(self._handle_execute_errors)
        self.session.hooks["response"].append(self._handle_api_errors)

    """ Хук для обработки множественных ошибок, возникающих в ходе выполнения метода `execute`"""
    def _handle_execute_errors(self, r, *args, **kwargs):
        try:
            errors = r.json()["execute_errors"]
            err_pairs = [(err["error_code"], err["error_msg"]) for err in errors]
            for err_code, err_msg in err_pairs:
                if self.api_errors[err_code]["action"] == "execute_break":
                    r.status_code = self.api_errors[err_code]["MatchedHTTPError"]
                    r.reason = err_msg
                    logger.debug(f"An error occured while processing `execute`")
                    logger.debug(f"Code has been changed to {self.api_errors[err_code]['MatchedHTTPError']}")
                    return r
            for error_code, error_msg in err_pairs:
                if self.api_errors[err_code]["action"] == "execute_retry":
                    r.status_code = self.api_errors[err_code]["MatchedHTTPError"]
                    r.reason = err_msg
                    logger.debug(f"An error occured while processing `execute`")
                    logger.debug(f"Code has been changed to {self.api_errors[err_code]['MatchedHTTPError']}")
                    return r
            for error_code, error_msg in err_pairs:
                if self.api_errors[err_code]["action"] == "execute_skip":
                    r.status_code = self.api_errors[err_code]["MatchedHTTPError"]
                    r.reason = err_msg
                    logger.debug(f"An error occured while processing `execute`")
                    logger.debug(f"Code has been changed to {self.api_errors[err_code]['MatchedHTTPError']}")
                    return r
            for error_code, error_msg in err_pairs:
                if self.api_errors[err_code]["action"] in ("skip", "breake", "retry"):
                    return r
            assert False, f"Unexpected error: {err_code}"
        except KeyError as e:
            return r

        

    """ Хук для обработки одиночных ошибок, возникающих при выполнении обычных методов """
    def _handle_api_errors(self, r, *args, **kwargs):
        try:
            err_code = r.json()["error"]["error_code"]
            err_msg = r.json()["error"]["error_msg"]
            if err_code not in self.api_errors:
                assert False, "Unexpected error code from VKApi"
            r.status_code = self.api_errors[err_code]['MatchedHTTPError']
            logger.debug(f"Code has been changed to {self.api_errors[err_code]['MatchedHTTPError']}")
            r.reason = err_msg
        except (AttributeError, KeyError) as e:
            pass
        finally:
            return r

    """ Вызов метода `execute`"""
    def execute(self, code:str):
        backoff_factor = 0.09
        params = {
             "access_token": self.access_token
            ,"code" : code
            ,"func_v": 1
            ,"v": "5.154"
        }
        method = "execute"
        logger.info(f"Run `{method}`")
        logger.debug(f"Query: {code}")
        r = self._retry_wrapper(method, params, backoff_factor)
        if r is None:
            return []
        return r.json()["response"]

    def get_users_ids(self, *args,  **kwargs) -> list[int]:
        try:
            friends = self.users_get(*args, **kwargs)
            return [user["id"] for user in friends]
        except KeyError as e:
            logger.debug(friends)
            logger.exception("Unexpected keys in response. `id` are Expected")
            raise SystemExit("Unexpected keys in response. `id` are Expected") from e

    """ Вызов метода `users.get`"""
    def users_get(self, screen_names:list[str], **kwargs) -> dict:
        backoff_factor = 0.09
        params = {
             "access_token": self.access_token
            ,"user_ids" : ",".join(map(str,screen_names))
            ,"v": "5.154"
            ,**kwargs
        }
        method = "users.get"
        logger.info(f"Run `{method}` for {screen_names}")
        r = self._retry_wrapper(method, params, backoff_factor)
        if r is None:
            logger.exception("Response object is None")
            raise SystemExit("Response object is None")
        try:
            return r.json()["response"]
        except KeyError as e:
            logger.debug(r.json())
            logger.exception("Unexpected keys in response. `response` are Expected")
            raise SystemExit("Unexpected keys in response. `response` are Expected") from e


    """ Вызов метода `friends.get`"""
    def friends_get(self, uid:int, **kwargs) -> dict[str,Any]:
        backoff_factor = 0.09
        params = {
             "access_token": self.access_token
            ,"user_id" : uid
            ,"v": "5.154"
            ,**kwargs
        }
        method = "friends.get"
        logger.info(f"Run `{method}` for {uid}")
        r = self._retry_wrapper(method, params, backoff_factor)
        if r is None:
            return {}
        try:
            return r.json()["response"] 
        except KeyError as e:
            logger.debug(r.json())
            logger.exception("Unexpected key in response. `response` is Expected")
            raise SystemExit("Unexpected key in response. `response` is Expected") from e

    def get_friends(self, uid:int, **kwargs) -> list[int]:
        try:
            friends = self.friends_get(uid, **kwargs)["items"]
            logger.info(f"Count of friends for user `{uid}` = {len(friends)}")
            return friends
        except KeyError as e:
            logger.debug(friends)
            logger.exception("Unexpected keys in response. `response.items` are Expected")
            raise SystemExit("Unexpected keys in response. `response.items` are Expected") from e

    def _retry_wrapper(self, method, params, backoff_factor=0.1):
        retries = 0
        url = urllib.parse.urljoin(self.api_endpoint, method)
        success = False
        r = None
        with self.session as s:
            while not success:
                try:
                    r = s.get(url, params = params)
                    r.raise_for_status()
                    success = True
                except requests.exceptions.HTTPError as err:
                    logger.info(f"Response api code: {err.response.status_code}")
                    retries += 1
                    if err.response.status_code in [self.api_errors[er]["MatchedHTTPError"] for er in self.api_errors if self.api_errors[er]["action"] == "skip"]:
                        logger.info(f"Skip -- `{err.response.reason}`")
                        return None
                    if err.response.status_code in [self.api_errors[er]["MatchedHTTPError"] for er in self.api_errors if self.api_errors[er]["action"] == "break"]:
                        logger.info(f"Break -- `{err.response.reason}`")
                        raise SystemExit
                    if err.response.status_code in [self.api_errors[er]["MatchedHTTPError"] for er in self.api_errors if self.api_errors[er]["action"] == "retry"]:
                        logger.info(f"Retry -- `{err.response.reason}`")
                        wait = backoff_factor * (2 ** (retries))
                        logger.debug(f"Wait {wait} seconds")
                        sys.stdout.flush()
                        time.sleep(wait)
                        continue
                    raise SystemExit(err)
            logger.debug(r.json())
            return r
