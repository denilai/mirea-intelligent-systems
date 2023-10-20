from typing import List, Any, Dict, Tuple, Optional, Set
from requests.adapters import HTTPAdapter
from urllib3.util import Retry
import requests
import os
import json
import sys
import time
import urllib.parse


from utils import create_logger, parse_config

logger = create_logger("VkAPI")



class VkApiAgent:

    
    def __init__(self, endpoint:str, access_token:str):
        dir_path = os.path.dirname(os.path.realpath(__file__))
        self.api_errors = parse_config(os.path.join(dir_path, "api_errors.yaml"))

        self.retried_errors:Dict[int, int] = {6:429}
        self.skiped_errors:Dict[int, int] = {30:400}
        self.api_endpoint = endpoint
        self.access_token = access_token

        retries = Retry(
            total=3,
            backoff_factor=0.1,
            status_forcelist=[502, 503, 504, 429, 404],
            allowed_methods={'GET'},
        )
        self.session = requests.Session()
        self.session.hooks["response"].append(self.handle_api_errors)
        self.session.mount('https://', HTTPAdapter(max_retries=retries))


    def handle_api_errors(self, r, *args, **kwargs):
        #TODO:: add support of execute_errors (execute method)
        try:
            errors = r.json()["execute_errors"]
            err_pairs = [(err["error_code"], err["error_msg"]) for err in errors]

        except AttributeError as e:
            try:
                err_code = r.json()["error"]["error_code"]
                err_msg = r.json()["error"]["error_msg"]
                if err_code not in self.api_errors:
                    assert False, "Unexpected err_code"
                    return r
                r.status_code = self.api_errors[err_code]['MatchedHTTPError']
                r.reason = err_msg
            except AttributeError as e:
                pass
        finally:
            return r
        

    def execute(self, code:str):
        backoff_factor = 0.09
        params = {
             "access_token": self.access_token
            ,"code" : code
            ,"func_v": 1
            ,"v": "5.154"
        }
        method = "execute"
        logger.info(f"Run {method}`")
        r = self._retry_wrapper(method, params, backoff_factor)
        if r is None:
            return []
        return r.json()["response"]

    def get_users_ids(self, screen_names:list[str],  **kwargs) -> list[int]:
        backoff_factor = 0.09
        params = {
             "access_token": self.access_token
            ,"user_ids" : ",".join(map(str,screen_names))
            ,"v": "5.154"
            ,**kwargs
        }
        method = "users.get"
        logger.info(f"Run {method} for `{screen_names}`")
        r = self._retry_wrapper(method, params, backoff_factor)
        if r is None:
            return []
        try:
            friends = r.json()["response"]
            friends_uids = [user["id"] for user in friends]
            return friends_uids
        except KeyError as e:
            logger.debug(r.json())
            raise SystemExit("Unexpected keys in response. `response` are Expected") from e
        

    def get_friends(self, uid:int, **kwargs) -> list[int]:
        backoff_factor = 0.09
        params = {
             "access_token": self.access_token
            ,"user_id" : uid
            ,"v": "5.154"
            ,**kwargs
        }
        method = "friends.get"
        logger.info(f"Run {method} for `{uid}`")
        r = self._retry_wrapper(method, params, backoff_factor)
        if r is None:
            return []
        try:
            friends = r.json()["response"]["items"]
            logger.info(f"Count of friends for user `{uid}` = {len(friends)}")
            return friends
        except KeyError as e:
            logger.debug(r.json())
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
                    logger.debug(f"Response api code: {err.response.status_code}")
                    retries += 1
                    if err.response.status_code in [self.api_errors[er]["MatchedHTTPError"] for er in self.api_errors if self.api_errors[er]["action"] == "skip"]:
                        logger.debug(f"Skip -- `{err.response.reason}`")
                        return None
                    if err.response.status_code in [self.api_errors[er]["MatchedHTTPError"] for er in self.api_errors if self.api_errors[er]["action"] == "retry"]:
                        logger.debug(f"Retry -- `{err.response.reason}`")
                        wait = backoff_factor * (2 ** (retries))
                        logger.debug(f"Wait {wait} seconds")
                        sys.stdout.flush()
                        time.sleep(wait)
                        continue
                    raise SystemExit(err)
            print(r.json())
            return r
