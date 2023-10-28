from typing import Callable, Iterable, Any
from multiprocessing import Pool
from itertools import cycle, chain, starmap
from functools import partial, wraps
from utils import parse_config, create_logger, exec_cmd, batched, concat, copy_file_to_remote_host, truncate_file

import csv
import os
import json
import yaml
import logging

from vkapi import VkApiAgent
from neo4jdb import Neo4jAgent
from decorators import timer, debuger, balance_workload



logger = create_logger("Main")

dir_path = os.path.dirname(os.path.realpath(__file__))

config = parse_config("./config.yaml")
vk_api = config["vk_api"]
access_tokens = vk_api["access_tokens"]
endpoint = vk_api["endpoint"]
list_of_apps = [VkApiAgent(endpoint, access_token) for access_token in access_tokens]


def get_friends_pairs(root:int, leafs:Iterable[int]) -> list[tuple[int,int]]:
    """get_friends_pairs(1, [2,3,4,5]) = [(1,2),(1,3),(1,4),(1,5)]"""
    logger.info(f"Create {len(leafs)} frendships pairs for uid {root}")
    return list(zip(cycle([root]),leafs))


def write_pairs_into_csv(filename, pairs):
    """Записать пары в файл"""
    logger.info(f"Write {len(pairs)} friednships pairs into file {filename}")
    with open(filename, "a") as f:
        spamwriter = csv.writer(f,delimiter = ",")
        spamwriter.writerows(pairs)

@timer
@debuger
@balance_workload(list_of_apps)
def get_friendships_pairs_for_each_uid(api:VkApiAgent, uids:Iterable[int]) -> list[list[tuple[int,int]]]:
    """ С помощью метода `execute` составить пары,
    вида (id1, id2), (id1, id3), (id1,...), где
    id1 -- рассматриваемый пользователь, а id2, id3
    и т.д. -- друзья id1
    Для каждого id создается список подобных пар
    :param api: VkApiAgent | Объект, инкапсулирующий
        обращения vk-приложения к API
    :param uids: Iterable[int] | Рассматриваемые id
    :return: Список пар (id, друг) для каждого id
    :rtype: list[list[tuple[int, int]]]
    """
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
    friends = concat(list(map(lambda x: x[1], root_leafs_list)))
    list_of_friendship_pairs = list(map(lambda x: get_friends_pairs(x[0], x[1]), root_leafs_list))
    return list_of_friendship_pairs



@timer
@debuger
@balance_workload(list_of_apps)
def get_friends_for_each_uid(api:VkApiAgent, uids: Iterable[int]) -> list[list[int]]:
    """ С помощью метода `execute` составить cписок
    друзей для каждого рассматриваемого пользователя
    :param api: VkApiAgent | Объект, инкапсулирующий
        обращения vk-приложения к API
    :param uids: Iterable[int] | Рассматриваемые id
    :return: Список друзей для каждого id
    :rtype: list[list[int]]
    """
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

 
@timer
@debuger
@balance_workload(list_of_apps)
def get_users_ids(api:VkApiAgent,uids:Iterable[int]):
    return api.get_users_ids(uids)

def process_uids_execute_version(config) -> list[tuple[int, int]]:
    classmates_screen_names:list[str] = config["input"]["classmates_raw"]

    logger.info(f"Classmates: {classmates_screen_names} ({len(classmates_screen_names)} elements)")

    classmates_uids:set[int] = set(get_users_ids(classmates_screen_names))
    logger.info(f"Classmates IDs: {classmates_uids} ({len(classmates_uids)} elements)")

    # уплощение вложенных списков
    # друзья одногруппников
    fr_1_lvl:set[int] = set(chain(*get_friends_for_each_uid(classmates_uids)))
    logger.info(f"Find friends of first level (friends of classmates, len = {len(fr_1_lvl)})")

    # уплощение вложенных списков
    # друзья друзей одногруппников
    fr_2_lvl:set[int] = set(chain(*get_friends_for_each_uid(fr_1_lvl)))
    logger.info(f"Find friends of second level (friends of friends of classmates, len = {len(fr_2_lvl)})")

    # друзья друзей друзей одногруппников
    fr_3_lvl:set[int] = set(chain(*get_friends_for_each_uid(fr_2_lvl)))
    logger.info(f"Find friends of third level (friends of friends of friends of classmates, len = {len(fr_3_lvl)})")

    # Максимальное количество пользователей в графе. Это множетсво будет использовано
    # для отбрасывания лишних связей при обогaщении друзьями 3 уровня
    unique_uids:set[int] = classmates_uids.union(fr_1_lvl).union(fr_2_lvl)
    logger.info(f"Count of unique IDs == {len(unique_uids)}")

    # Пользователи, для которых будут найдены все друзья
    uids_to_process:set[int] = classmates_uids.union(fr_1_lvl)
    
    # пары, которыми будет обогащен основной набор
    additional_pairs:list[tuple[int,int]] = concat(get_friendships_pairs_for_each_uid(fr_2_lvl))

    # из данного списка пар исключены узлы, которые отсутствуют в `unique_uids`
    filtered_pairs:list[tuple[int,int]] = list(filter(lambda x: x[1] in unique_uids, additional_pairs))

    #list_of_friendship_pairs:list[list[tuple[int, int]]] = get_friendships_pairs_for_each_uid(unique_uids)
    # пары друзей для множества `uids_to_process`
    friendship_pairs:list[tuple[int,int]] = concat(get_friendships_pairs_for_each_uid(uids_to_process))

    all_pairs = set(additional_pairs + friendship_pairs)

    return all_pairs


if __name__ == "__main__":
    # Считываение конфигируационного файла
    config = parse_config("./config.yaml")

    # Файл, в который будут записаны пары вида (<id>, <friend_id>)
    csv_file = config["result"]["csv_file"] 
    neo4j_import_dir = config["neo4j"]["import_dir"]
    user = config["remote_host"]["user"]
    host = config["remote_host"]["host"]

    src  = os.path.join(dir_path, csv_file)
    dest = os.path.join(neo4j_import_dir, csv_file) 

    list_of_friendship_pairs:list[tuple[int, int]] = concat(process_uids_execute_version(config))
    print(list_of_friendship_pairs)

    truncate_file(csv_file)

    write_pairs_into_csv(csv_file, list_of_friendship_pairs)

    copy_file_to_remote_host(src, dest, host, user)
    
    neo4j = config["neo4j"]

    neo4j_db = Neo4jAgent(
          neo4j["host"]
        , neo4j["port"]
        , neo4j["user"]
        , neo4j["password"]
        , neo4j["database"]
    )
    neo4j_db.detach_delete_persons()
    neo4j_db.load_person_from_csv(csv_file)
    neo4j_db.close()
