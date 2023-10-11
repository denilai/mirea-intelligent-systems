from typing import List, Any, Dict, Tuple, Optional, Set
import requests
import urllib.parse
import multiprocessing
mgr = multiprocessing.Manager()
sem = mgr.Semaphore(1)


class VkApiAgent:
    def __init__(self, endpoint:str, access_token:str):
        self.api_endpoint = endpoint
        self.access_token = access_token

    def get_friends(self, user_id, **kwargs) -> List[int]:
        import json
        params = {
             "access_token": self.access_token
            ,"user_id" : user_id
            ,"v": "5.154"
            ,**kwargs
        }
        method = "friends.get"
        url = urllib.parse.urljoin(self.api_endpoint, method)
        #with sem:
        try:
            r = requests.get(url, params = params)
            r.raise_for_status()

        except requests.exceptions.HTTPError as err:
            raise(SystemExit(err))
        try:
            return r.json().get("response").get("items")
        except AttributeError as e:
          
          return []
        except KeyError as e:
            raise SystemExit("Unexpected keys in response. `response.items` are Expected") from e
            return []
