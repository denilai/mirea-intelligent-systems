from typing import List, Any, Dict, Tuple, Optional, Set
from neo4j import GraphDatabase
from itertools import cycle

import csv
import requests
import urllib.parse
import json
import yaml


logging.basicConfig(level = logging.INFO)
logger = logging.getLogger("vk")


def parse_config(filepath:str) -> Dict[str, Any]:
  with open(filepath,"r") as f:
    config = yaml.safe_load(f)
  return config


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

    def _load_person_from_csv(self, tx, filepath):
      query = f"""load csv from $filepath as line
      merge (p1:Person {id:line[0]})
      merge (p2:Person {id:line[1]})
      merge (p1)-[:IS_FRIENDS_WITH]->(p2);"""
      params = {"filepath":filepath}
      tx.run(query, parameters = params)


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

    def _add_person(self, tx, person_id):
        result = tx.run("CREATE (p: Person {id: $person_id})", person_id = person_id)

    def add_person_friends(self, root_id, leaf_ids:List[str]) -> None:
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._add_person_friends,root_id, leaf_ids)


    def load_person_from_csv(self, filepath):
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._load_person_from_csv, filepath)

    def add_person(self, person_id):
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._add_person, person_id)
            print(f"Person with id = {person_id} was created") 


def create_friends_pairs(root:int, leafs:int) -> List[Tuple[str,str]]:
  """Create pairs [(x,y0),(x,y1),(x,y2)..] from x:str and ys:List[str] """
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

  uid_to_process:Set[int] = set(groupmates)
  for uid in groupmates: 
    uid_to_process = uid_to_process.union(api_instance.get_friends(uid))

  with open(config["result"]["csv_file"], "w") as f:
    for uid in uid_to_process:
      process_user(api_instance, uid, f)

  neo4j_db.close()


def process_user(api_instance, uid:int, f) -> List[Tuple[str,str]]:
  """Find user friends and write pairs `(user_id,friend_id)` to csv file"""
  friends:List[int] = api_instance.get_friends(uid)
  logger.info(f"Count friends of user {uid} = {len(friends)}")
  f_pairs:List[Tuple[str, str]] = create_friends_pairs(uid, friends)
  write_pairs_into_csv(f, f_pairs)


if __name__ == "__main__":
  config = parse_config("./config.yaml")
  create_friend_graph(config)
