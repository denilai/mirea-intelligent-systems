from typing import List, Any, Dict, Tuple, Optional, Set, Callable, Iterable
from multiprocessing import Pool, TimeoutError
from itertools import cycle, chain
from functools import partial, reduce
from utils import parse_config, create_logger, exec_cmd

import csv
import os
import json
import yaml
import logging

from vkapi import VkApiAgent
from neo4jdb import Neo4jAgent




dir_path = os.path.dirname(os.path.realpath(__file__))

def create_friends_pairs(root:int, leafs:list[int]) -> List[Tuple[str,str]]:
    """Create pairs [(x,y0),(x,y1),(x,y2)..] from x:str and ys:List[str] """
    return list(zip(cycle([str(root)]), map(lambda x: str(x),leafs)))

def write_pairs_into_csv(logger, filename, pairs):
    with open(filename, "a") as f:
        spamwriter = csv.writer(f,delimiter = ",")
        logger.info("Write into file")
        print(f)
        spamwriter.writerows(pairs)

def process_user(api_instance, uid:int) -> list[Tuple[str, str]]:
    """Find user friends and write pairs `(user_id,friend_id)` to csv file"""
    friends:List[int] = api_instance.get_friends(uid)
    #logger.info(f"Count friends of user {uid} = {len(friends)}")
    f_pairs:List[Tuple[str, str]] = create_friends_pairs(uid, friends)
    return f_pairs

        #res = pool.map(api_instance.get_friends,xs

def parallel_runner(f:Callable, xs:Iterable, iter_type, processes:int = 5):
    with Pool(processes=processes) as pool:
        res = pool.map(f,xs)
    return iter_type(chain(*res))

def get_friends_pairs(logger, config):
    vk_api = config["vk_api"]
    groupmates_screen_names = config["input"]["groupmates_raw"]
    api_instance = VkApiAgent(vk_api["endpoint"], vk_api["access_token"])

    groupmates:Set[int] = set([user["id"] for user in api_instance.get_users(groupmates_screen_names)])

    unique_uid = parallel_runner(api_instance.get_friends, groupmates, set, processes = 10)
    logger.info(f"Count ot unique UID = {len(unique_uid)}")
    friends_pairs = parallel_runner(partial(process_user, api_instance),unique_uid, list, processes = 10)

    #uid_to_process = reduce(lambda x,y: x.union(api_instance.get_friends(y)), uid_to_process, uid_to_process)

    #with Pool(processes=5) as pool:
    #    res = pool.map(partial(process_user,api_instance), uid_to_process)
    #friends_pairs = list(chain(*res))
    logging.info(f"All UIDs have been processed. Count of friends relationships = {len(friends_pairs)}")
    return friends_pairs


def copy_file_to_remote_host(logger, src, dest, host, user):
    src  = os.path.join(dir_path, csv_file)
    dest = os.path.join(neo4j_import_dir, csv_file) 
    cmd  = f"scp {src} {user}@{host}:{dest}"
    logger.info(f"Copy {csv_file} to datastore.lab.denisov")
    exec_cmd(cmd.split(" "), logger)


if __name__ == "__main__":
    logger = create_logger("VK")
    config = parse_config("./config.yaml")

    csv_file = config["result"]["csv_file"] 
    neo4j_import_dir = config["neo4j"]["import_dir"]
    user = config["remote_host"]["user"]
    host = config["remote_host"]["host"]


    src  = os.path.join(dir_path, csv_file)
    dest = os.path.join(neo4j_import_dir, csv_file) 

    friends_pairs = get_friends_pairs(logger, config)
    write_pairs_into_csv(logger, csv_file, friends_pairs)
    copy_file_to_remote_host(logger, src, dest, host, user)
    

    neo4j = config["neo4j"]
    neo4j_db = Neo4jAgent(neo4j["host"], neo4j["port"], neo4j["user"], neo4j["password"], neo4j["database"])
    neo4j_db.load_person_from_csv(csv_file)

    neo4j_db.close()
