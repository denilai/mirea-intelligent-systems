from typing import List, Any, Dict, Tuple, Optional
from neo4j import GraphDatabase

import time
import requests
import urllib.parse
import json
import yaml


def parse_config(filepath:str) -> Dict[str, Any]:
    with open(filepath,"r") as f:
        config = yaml.safe_load(f)
    return config

    

class VkApiAgent:
    def __init__(self, endpoint:str, access_token:str):
        self.api_endpoint = endpoint
        self.access_token = access_token

    def get_friends(self, user_id, **kwargs) -> List[str]:
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
        except KeyError as e:
            raise SystemExit("Unexpected keys in response. `response.items` are Expected") from e
            return []

class GraphDb:
    def __init__(self, neo4j_endpoint:str, neo4j_user:str, neo4j_pass:str, neo4j_db:str):
        self.driver = GraphDatabase.driver(neo4j_endpoint, auth=(neo4j_user, neo4j_pass))
        self.database = neo4j_db

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
        #print(query)
        #print(params)
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


        
def create_friend_graph(config):
    print(config)
    exit(1)
    try:
        neo4j = config["neo4j"]
        vk_api = config["vk_api"]
        groupmates = config["input"]["groupmates"]
    except KeyError as e:
        raise Exception("Wrong cofnig. Fileds `neo4j`, `input.groupmates` and `vk_api` are expected") from e
    print(config)
    neo4j_db = GraphDb(neo4j["endpoint"], neo4j["user"], neo4j["pass"], neo4j["database"])
    api_instance = VkApiAgent(vk_api["endpoint"], vk_api["access_token"])
    for user_id in groupmates:
        root_friends = list(map(lambda x: str(x),api_instance.get_friends(user_id)))
        #root_friends = ["185362653"]
        print(f"Count friends of user {user_id} = {len(root_friends)}")
        neo4j_db.add_person_friends(user_id, root_friends)
        if len(root_friends) > 300:
            for friend_id in root_friends:
                friend_friends = list(map(lambda x: str(x),api_instance.get_friends(friend_id)))
                print(f"Count friends of user {friend_id} = {len(friend_friends)}")
                if len(friend_friends) < 700:
                    neo4j_db.add_person_friends(friend_id, friend_friends)

        #print(root_friends)
        #query = neo4j_conn.create_friends_subgraph(user, root_friends)

    #neo4j_conn.visualize(graph_vizualization_filename)
    neo4j_db.close()



if __name__ == "__main__":
    config = parse_config("./config.yaml")
    create_friend_graph(config)





