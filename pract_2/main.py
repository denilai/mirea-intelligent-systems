from typing import List, Any, Dict, Tuple, Optional, Set, Callable, Iterable
from multiprocessing import Pool, TimeoutError
from itertools import cycle, chain
from functools import partial, reduce
from utils import parse_config, create_logger, exec_cmd, batched

import csv
import os
import json
import yaml
import logging

from vkapi import VkApiAgent
from neo4jdb import Neo4jAgent



logger = create_logger("Main")



dir_path = os.path.dirname(os.path.realpath(__file__))

def get_friends_pairs(root:int, leafs:list[int]) -> List[Tuple[str,str]]:
    """Create pairs [(x,y0),(x,y1),(x,y2)..] from x:str and ys:List[str] """
    logger.info(f"Create {len(leafs)} frendships pairs for uid {root}")
    return list(zip(cycle([str(root)]), map(lambda x: str(x),leafs)))

def write_pairs_into_csv(filename, pairs):
    logger.info(f"Write {len(pairs)} friednships pairs into file {filename}")
    with open(filename, "a") as f:
        spamwriter = csv.writer(f,delimiter = ",")
        spamwriter.writerows(pairs)

def process_user(api_instance, uid:int) -> list[Tuple[str, str]]:
    """Find user friends and write pairs `(user_id,friend_id)` to csv file"""
    friends:List[int] = api_instance.get_friends(uid)
    #logger.info(f"Count friends of user {uid} = {len(friends)}")
    f_pairs:List[Tuple[str, str]] = get_friends_pairs(uid, friends)
    return f_pairs

        #res = pool.map(api_instance.get_friends,xs

def parallel_runner(f_map:Callable, xs:Iterable, f_out:Callable, processes:int = 5):
    logger.info(f"Running func {f_map} in {processes} processes. Out func: {f_out.__name__} with list {xs}")
    with Pool(processes=1) as pool:
        res = pool.map(f_map,xs)
    return f_out(res)

def get_friends_pair2s(config, api_instance):
    classmates_screen_names = config["input"]["classmates_raw"]

    classmates:Set[int] = set([user["id"] for user in api_instance.get_users(classmates_screen_names)])

    unique_uid = classmates.union(parallel_runner(api_instance.get_friends, classmates, lambda x: set(chain(*x)), processes = 10))
    logger.info(f"Count ot unique UID = {len(unique_uid)}")
    friends_pairs = parallel_runner(partial(process_user, api_instance),unique_uid, lambda x: list(chain(*x)), processes = 10)

    #uid_to_process = reduce(lambda x,y: x.union(api_instance.get_friends(y)), uid_to_process, uid_to_process)

    #with Pool(processes=5) as pool:
    #    res = pool.map(partial(process_user,api_instance), uid_to_process)
    #friends_pairs = list(chain(*res))
    logging.info(f"All UIDs have been processed. Count of friends relationships = {len(friends_pairs)}")
    return friends_pairs


def get_list_of_friendship_pairs(api, uids:list[int]) -> list[Tuple[int, int]]:
    logger.info(f"Get list of friendship pairs for {uids}")
    code = f"""var b = {list(uids)};
    var count = b.length;
    var i = 0;
    var res = [];
    while (i<count) {{
        var friends = API.friends.get({{"user_id":b[i]}});
        if (friends.items == null) {{
            res.push([b[i],[]]);
        }}
        else {{
            res.push([b[i],friends.items]);
        }}
        i=i+1;
    }};

    return res;"""
    root_leafs_list = api.execute(code)
    list_of_friendship_pairs = list(map(lambda x: get_friends_pairs(x[0], x[1]), root_leafs_list))
    return list_of_friendship_pairs


def get_list_of_friends(api,uids: list[int]) -> list[list[int]]:
    code = f"""var b = {list(uids)};
    var count = b.length;
    var i = 0;
    var res = [];
    while (i<count) {{
        var friends = API.friends.get({{"user_id":b[i]}});
        if (friends.items == null) {{
            res.push([]);
        }}
        else {{
            res.push(friends.items);
        }}
        i=i+1;
    }};
    return res;"""
    return api.execute(code)

def process_uids_execute_version(config, api):
    csv_file = config["result"]["csv_file"] 
    classmates_screen_names = config["input"]["classmates_raw"]

    logger.info(f"Input user names: {classmates_screen_names}")

    classmates_uids:Set[int] = set([user["id"] for user in api_instance.get_users(classmates_screen_names)])
    logger.info(f"Input user ids: {classmates_uids}")

    # flattening nested lists
    friends_of_classmates = set(chain(*get_list_of_friends(api, classmates_uids)))
    unique_uids = classmates_uids.union(friends_of_classmates)

    logger.info(f"Count of unique uids to be processed: {len(unique_uids)}")

    # split list to chunks to avoid api overloading
    splitted_unique_uids = list(batched(unique_uids, 20))

    # running func `get_list_of_friendships_pairs` in parallel and processing its result
    list_of_friendship_pairs = parallel_runner(
         partial(get_list_of_friendship_pairs, api)
        ,splitted_unique_uids
        ,lambda z: list(reduce(lambda x,y: x+y, z, []))
        ,10
    )
    return list_of_friendship_pairs
    
    

def copy_file_to_remote_host(src, dest, host, user):
    src  = os.path.join(dir_path, csv_file)
    dest = os.path.join(neo4j_import_dir, csv_file) 
    cmd  = f"scp {src} {user}@{host}:{dest}"
    logger.info(f"Copy {csv_file} to datastore.lab.denisov")
    exec_cmd(cmd.split(" "), logger)

def truncate_file(filename):
    logger.info(f"Truncate file {filename}")
    open(filename, "w")
    return
    



if __name__ == "__main__":
    config = parse_config("./config.yaml")

    csv_file = config["result"]["csv_file"] 
    neo4j_import_dir = config["neo4j"]["import_dir"]
    user = config["remote_host"]["user"]
    host = config["remote_host"]["host"]
    vk_api = config["vk_api"]

    api_instance = VkApiAgent(vk_api["endpoint"], vk_api["access_token"])

    src  = os.path.join(dir_path, csv_file)
    dest = os.path.join(neo4j_import_dir, csv_file) 

    list_of_friendship_pairs = process_uids_execute_version(config, api_instance)
    truncate_file(csv_file)
    for pairs in list_of_friendship_pairs:
        write_pairs_into_csv(csv_file, pairs)
    copy_file_to_remote_host(src, dest, host, user)
    
    neo4j = config["neo4j"]
    neo4j_db = Neo4jAgent(neo4j["host"], neo4j["port"], neo4j["user"], neo4j["password"], neo4j["database"])
    neo4j_db.detach_delete_persons()
    neo4j_db.load_person_from_csv(csv_file)

    neo4j_db.close()
