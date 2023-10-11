from typing import List, Any, Dict, Tuple, Optional, Set
from multiprocessing import Pool, TimeoutError
from itertools import cycle, chain
from functools import partial

import csv
import json
import yaml
import logging

from vkapi import VkApiAgent
from neo4jdb import Neo4jAgent


logging.basicConfig(level = logging.INFO)
logger = logging.getLogger("vk")



def create_friends_pairs(root:int, leafs:int) -> List[Tuple[str,str]]:
    """Create pairs [(x,y0),(x,y1),(x,y2)..] from x:str and ys:List[str] """
    return list(zip(cycle([str(root)]), map(lambda x: str(x),leafs)))

def write_pairs_into_csv(filename, pairs):
    with open(filename, "a") as f:
        spamwriter = csv.writer(f,delimiter = ",")
        logger.info("Write into file")
        print(f)
        spamwriter.writerows(pairs)

def process_user(api_instance, uid:int):
    """Find user friends and write pairs `(user_id,friend_id)` to csv file"""
    friends:List[int] = api_instance.get_friends(uid)
    logger.info(f"Count friends of user {uid} = {len(friends)}")
    f_pairs:List[Tuple[str, str]] = create_friends_pairs(uid, friends)
    return f_pairs
                
def parse_config(filepath:str) -> Dict[str, Any]:
    with open(filepath,"r") as f:
        config = yaml.safe_load(f)
    return config

def create_friend_graph(config):
    try:
        neo4j = config["neo4j"]
        vk_api = config["vk_api"]
        groupmates = config["input"]["groupmates"]
    except KeyError as e:
        raise SystemExit("Wrong cofnig. Fileds `neo4j`, `input.groupmates` and `vk_api` are expected") from e
    api_instance = VkApiAgent(vk_api["endpoint"], vk_api["access_token"])
    neo4j_db = Neo4jAgent(neo4j["host"], neo4j["port"], neo4j["user"], neo4j["password"], neo4j["database"])
    
    uid_to_process:Set[int] = set(groupmates)

    for uid in groupmates: 
        uid_to_process = uid_to_process.union(api_instance.get_friends(uid))
    #uid_to_process = [133329982, 228620383, 200372810]
    filename = config["result"]["csv_file"] 
    with Pool(processes=10) as pool:
        res = pool.map(partial(process_user,api_instance), uid_to_process)
        print(res)

    lists = list(chain(*res))
    print(len(lists))
    logging.info("All uids have been processed")

    neo4j_db.close()




if __name__ == "__main__":
    config = parse_config("./config.yaml")
    create_friend_graph(config)
