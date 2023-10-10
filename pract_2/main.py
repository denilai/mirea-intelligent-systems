from typing import List, Any, Dict, Tuple, Optional
from neo4j import GraphDatabase
from itertools import cycle

import csv
import time
import requests
import urllib.parse
import json
import yaml

from utils import parse_config
from pg import logger


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



class Neo4jAgent:
    def __init__(self,
                 host:str,
                 port:str,
                 user:str,
                 password:str,
                 database:str
                 ):
        endpoint = f"bolt://{host}:{port}"
        self.driver = GraphDatabase.driver(endpoint, auth=(user,password))
        self.database = database

    def close(self):
        self.driver.close()

    def _add_person_friends(self, tx, root_id:str, leaf_ids:List[str]) -> None:
        query = f"""merge (root:Person {{id:$root_id}}) """
        params = {"root_id":root_id}
        for i, leaf_id  in enumerate(leaf_ids):
            query += f"""
            merge (leaf_{i}:Person {{id:$leaf_{i}id}})
            merge (root)-[:IS_FRIENDS_WITH]->(leaf_{i})
            """
            params[f"leaf_{i}id"]=leaf_id
        tx.run(query, parameters = params)

    def add_person_friends(self, root_id, leaf_ids:List[str]) -> None:
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._add_person_friends,root_id, leaf_ids)


    def add_person(self, person_id):
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._add_person, person_id)
            print(f"Person with id = {person_id} was created") 

    def _add_person(self, tx, person_id):
        result = tx.run("CREATE (p: Person {id: $person_id})", person_id = person_id)



def create_friends_pairs(root:int, leafs:int) -> List[Tuple[str,str]]:
  return list(zip(cycle([str(root)]), map(lambda x: str(x),leafs)))

def write_pairs_into_csv(f, pairs):
  spamwriter = csv.writer(f,delimiter = ",")
  spamwriter.writerows(pairs)

        
def create_friend_graph(config):
  try:
    neo4j = config["neo4j"]
    vk_api = config["vk_api"]
    groupmates = config["input"]["groupmates"]
  except KeyError as e:
    raise SystemExit("Wrong cofnig. Fileds `neo4j`, `input.groupmates` and `vk_api` are expected") from e
  neo4j_db = Neo4jAgent(neo4j["host"], neo4j["port"], neo4j["user"], neo4j["password"], neo4j["database"])
  api_instance = VkApiAgent(vk_api["endpoint"], vk_api["access_token"])
  with open(config["result"]["csv_file"], "w") as f:
    #a = [process_user(api_instance, friend, f) for friend in [process_user(api_instance, user, f) in groupmates]]
    #a = [process_user(api_instance, friend, f) for friend in [process_user(api_instance, user, f) for user in groupmates]]
    a = [[list(map(lambda friend: process_user(api_instance, friend, f), friends)) for friends in [process_user(api_instance, user, f) for user in groupmates]]]
  print(len(a))





def process_user(api_instance, uid:int, f) -> List[int]:
  friends:List[int] = api_instance.get_friends(uid)
  logger.info(f"Count friends of user {uid} = {len(friends)}")
  f_pairs:List[Tuple[str, str]] = create_friends_pairs(uid, friends)
  write_pairs_into_csv(f, f_pairs)
  logger.info("Write pairs into {f}")
  return friends




if __name__ == "__main__":
    config = parse_config("./config.yaml")
    create_friend_graph(config)





