from typing import Callable, Iterable
from multiprocessing import Pool
from itertools import cycle, chain, starmap
from functools import partial
from utils import parse_config, create_logger, exec_cmd, batched, concat, copy_file_to_remote_host, truncate_file

import csv
import os
import json
import yaml
import logging

from vkapi import VkApiAgent
from neo4jdb import Neo4jAgent


# count of IDs for processing in `execute` method of VkAPI
API_LIST_THRESHOLD = 20


logger = create_logger("Main")

dir_path = os.path.dirname(os.path.realpath(__file__))

def get_friends_pairs(root:int, leafs:list[int]) -> list[tuple[int,int]]:
    """Create pairs [(x,y0),(x,y1),(x,y2)..] from x:str and ys:list[str] """
    logger.info(f"Create {len(leafs)} frendships pairs for uid {root}")
    return list(zip(cycle([root]),leafs))


def write_pairs_into_csv(filename, pairs):
    logger.info(f"Write {len(pairs)} friednships pairs into file {filename}")
    with open(filename, "a") as f:
        spamwriter = csv.writer(f,delimiter = ",")
        spamwriter.writerows(pairs)



def get_friendships_pairs_for_each_uid(api, uids:Iterable[int]) -> list[list[tuple[int,int]]]:
    logger.info(f"Search friendship pairs for {uids}")
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


def get_friends_for_each_uid(api,uids: Iterable[int]) -> list[list[int]]:
    logger.info(f"Search friends for {uids}")
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


def four_lvl(api, uids:Iterable[int]) -> list[list[tuple[int,int]]]:
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


def process_uids_execute_version(config, api) -> list[list[tuple[int, int]]]:
    csv_file = config["result"]["csv_file"] 
    classmates_screen_names:list[str] = config["input"]["classmates_raw"]

    logger.info(f"Classmates: {classmates_screen_names} ({len(classmates_screen_names)} elements)")

    classmates_uids:set[int] = set(api_instance.get_users_ids(classmates_screen_names))
    logger.info(f"Classmates IDs: {classmates_uids} ({len(classmates_uids)} elements)")

    # flattening nested lists
    logger.info("Find friends of first level (friends of classmates)")
    fr_1_lvl:set[int] = set(chain(*get_friends_for_each_uid(api, classmates_uids)))

    # split list to chunks to avoid api overloading
    fr_1_lvl_s:list[set[int]] = list(batched(fr_1_lvl, API_LIST_THRESHOLD))
    logger.info(f"Split list of IDs about {len(fr_1_lvl)} to {len(fr_1_lvl_s)} chuncks about {API_LIST_THRESHOLD} elements")

    # flattening nested lists
    logger.info("Find friends of second level (friends of friends of classmates)")
    fr_2_lvl:set[int] = set(chain(*list(map(lambda x: set(chain(*get_friends_for_each_uid(api, x))), fr_1_lvl_s))))

    unique_uids:set[int] = classmates_uids.union(fr_1_lvl).union(fr_2_lvl)
    logger.info(f"Count of unique IDs == {len(unique_uids)}")

    # split list to chunks to avoid api overloading
    splitted_unique_uids:list[tuple] = list(batched(unique_uids, API_LIST_THRESHOLD))
    logger.info(f"Split list of IDs about {len(unique_uids)} to {len(splitted_unique_uids)} chuncks about {API_LIST_THRESHOLD} elements")

    list_of_friendship_pairs:list[list[tuple[int, int]]] = concat(
          starmap(
              get_friendships_pairs_for_each_uid
            , list(
                 zip(cycle([api])
               , splitted_unique_uids)
            )
        )
    )
    exit(1)
    return list_of_friendship_pairs


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

    list_of_friendship_pairs:list[tuple[int, int]] = concat(process_uids_execute_version(config, api_instance))
    print(list_of_friendship_pairs)

    truncate_file(csv_file)

    write_pairs_into_csv(csv_file, list_of_friendship_pairs)

    copy_file_to_remote_host(src, dest, host, user)
    
    neo4j = config["neo4j"]

    neo4j_db = Neo4jAgent(neo4j["host"], neo4j["port"], neo4j["user"], neo4j["password"], neo4j["database"])
    neo4j_db.detach_delete_persons()
    neo4j_db.load_person_from_csv(csv_file)

    neo4j_db.close()
