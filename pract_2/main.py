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


# Предел количества элементов,
# которые могут быть обработаны в одном методе `execute`
API_LIST_THRESHOLD = 20


logger = create_logger("Main")

dir_path = os.path.dirname(os.path.realpath(__file__))

config = parse_config("./config.yaml")
vk_api = config["vk_api"]
access_tokens = vk_api["access_tokens"]
endpoint = vk_api["endpoint"]
list_of_apps = [VkApiAgent(endpoint, access_token) for access_token in access_tokens]


def balance_workload(apis:list[VkApiAgent]) -> Callable:
    """ Функция-дeкоратор для разделения нагрузки между
    отдельными vk-приложениями.
    :param apis: list[VkApiAgent] Список объектов, инкапсулирующих
        обращение приложений к VK API
    :rtype: Callable
    """
    def inner(f:Callable[[VkApiAgent, list[Any]], list[Any]]) -> Callable:
        """
        :param f: Callable Захватываемая функция
        :rtype: Callable
        """
        @wraps(f)
        def wrapper(args:list[Any]) -> list[Any]:
            """
            :param args: list[Any] Cписок аргументов для вызова f
            """
            # разделяем список аргументов на части,
            # чтобы избежать превышения лимита по размеру тела ответа от API
            logger.debug(f"Args = {args}")
            splitted_args:list[list[Any]] = list(batched(args, API_LIST_THRESHOLD))
            logger.info(f"Split list about {len(args)} to {len(splitted_args)} chuncks about {API_LIST_THRESHOLD} elements")
            return concat([f(api,args) for api, args in zip(cycle(apis),splitted_args)])
        return wrapper
    return inner



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


@balance_workload(list_of_apps)
def get_friendships_pairs_for_each_uid(api, uids:Iterable[int]) -> list[list[tuple[int,int]]]:
    """ С помощью метода `execute` составить пары,
    вида (id1, id2), (id1, id3), (id1,...), где
    id1 -- рассматриваемый пользователь, а id2, id3
    и т.д. -- друзья id1
    Для каждого id создается список подобных пар
    :param api: VkApiAgent Объект, инкапсулирующий
        обращения vk-приложения к API
    :param uids: Iterable[int] Рассматриваемые id
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
    list_of_friendship_pairs = list(map(lambda x: get_friends_pairs(x[0], x[1]), root_leafs_list))
    return list_of_friendship_pairs


@balance_workload(list_of_apps)
def get_friends_for_each_uid(api, uids: Iterable[int]) -> list[list[int]]:
    """ С помощью метода `execute` составить cписок
    друзей для каждого рассматриваемого пользователя
    :param api: VkApiAgent Объект, инкапсулирующий
        обращения vk-приложения к API
    :param uids: Iterable[int] Рассматриваемые id
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


@balance_workload(list_of_apps)
def get_users_ids(api,uids):
    return api.get_users_ids(uids)

def process_uids_execute_version(config) -> list[list[tuple[int, int]]]:
    csv_file = config["result"]["csv_file"] 
    classmates_screen_names:list[str] = config["input"]["classmates_raw"]

    logger.info(f"Classmates: {classmates_screen_names} ({len(classmates_screen_names)} elements)")

    classmates_uids:set[int] = set(get_users_ids(classmates_screen_names))
    logger.info(f"Classmates IDs: {classmates_uids} ({len(classmates_uids)} elements)")

    # flattening nested lists
    fr_1_lvl:set[int] = set(chain(*get_friends_for_each_uid(classmates_uids)))
    logger.info(f"Find friends of first level (friends of classmates, len = {len(fr_1_lvl)})")

    # flattening nested lists
    fr_2_lvl:set[int] = set(chain(*get_friends_for_each_uid(fr_1_lvl)))
    logger.info(f"Find friends of second level (friends of friends of classmates, len = {len(fr_2_lvl)})")

    assert False, "Not implemented yet"
    unique_uids:set[int] = classmates_uids.union(fr_1_lvl).union(fr_2_lvl)
    logger.info(f"Count of unique IDs == {len(unique_uids)}")

    list_of_friendship_pairs:list[list[tuple[int, int]]] = concat(map(get_friendships_pairs_for_each_uid, unique_uids))
    exit(1)
    return list_of_friendship_pairs


if __name__ == "__main__":
    config = parse_config("./config.yaml")

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

    neo4j_db = Neo4jAgent(neo4j["host"], neo4j["port"], neo4j["user"], neo4j["password"], neo4j["database"])
    neo4j_db.detach_delete_persons()
    neo4j_db.load_person_from_csv(csv_file)

    neo4j_db.close()
