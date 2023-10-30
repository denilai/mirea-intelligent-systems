from typing import Callable, Iterable
from multiprocessing import Pool
from functools import wraps, partial
from typing import Any
from itertools import cycle, starmap as itertools_starmap
import time
from utils import create_logger, batched, API_LIST_THRESHOLD, concat
from vkapi import VkApiAgent
from multiprocessing import Pool, TimeoutError


logger = create_logger("Decor")


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
        #@wraps(f)
        def wrapper(args:list[Any]) -> list[Any]:
            """
            :param args: list[Any] Cписок аргументов для вызова f
            """
            logger.debug(f"Args = {args}")
            # разделяем список аргументов на части,
            # чтобы избежать превышения лимита по размеру тела ответа от API
            splitted_args:list[list[Any]] = list(batched(args, API_LIST_THRESHOLD))
            logger.info(f"Split list about {len(args)} to {len(splitted_args)} chuncks about {API_LIST_THRESHOLD} elements")
            res = [f(api,args) for api, args in zip(cycle(apis),splitted_args)]
            logger.debug(res)
            res = concat(res)
            logger.debug(f"Concatenated result: {res}")
            return res
        return wrapper
    return inner

def debuger(func):
    """Print the function signature and return value"""
    @wraps(func)
    def wrapper_debug(*args, **kwargs):
        args_repr = [repr(a) for a in args]                      # 1
        kwargs_repr = [f"{k}={v!r}" for k, v in kwargs.items()]  # 2
        signature = ", ".join(args_repr + kwargs_repr)           # 3
        logger.debug(f"Calling {func.__name__}({signature})")
        value = func(*args, **kwargs)
        logger.debug(f"{func.__name__!r} returned {value!r}")           # 4
        return value
    return wrapper_debug

def timer(func):
    """Print the runtime of the decorated function"""
    @wraps(func)
    def wrapper_timer(*args, **kwargs):
        start_time = time.perf_counter()    # 1
        value = func(*args, **kwargs)
        end_time = time.perf_counter()      # 2
        run_time = end_time - start_time    # 3
        logger.debug(f"Finished {func.__name__!r} in {run_time:.4f} secs")
        return value
    return wrapper_timer

def get_friends_pairs(root:int, leafs:list[int]) -> list[tuple[int,int]]:
    """Create pairs [(x,y0),(x,y1),(x,y2)..] from x:str and ys:list[str] """
    logger.info(f"Create {len(leafs)} frendships pairs for uid {root}")
    return list(zip(cycle([root]),leafs))


        


def main():
    def parallel_decorator(processes = 5):
        def wrapper(f):
            @wraps(f)
            def work(api, xs):
                with Pool(processes = processes) as pool:
                    res = pool.starmap(f, [(api, xs[0]), (api, xs[1])])
                return res
            return work
        return wrapper

    @parallel_decorator(processes=1)
    def test(x, y):
        print("X = {}, Y = {}".format(x, y))

    print(test(1,[1,2]))

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




if __name__ == "__main__":
    main()
    
