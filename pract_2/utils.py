from typing import Dict, Any, Callable, Tuple, Mapping
import yaml
import logging
import os
import subprocess as subproc
import re


_find_unsafe = re.compile(r'[^\w@%+=:,./-]', re.ASCII).search

def quote(s):
    """Return a shell-escaped version of the string *s*."""
    if not s:
        return "''"
    if _find_unsafe(s) is None:
        return s

    # use single quotes, and put single quotes into double quotes
    # the string $'b is then quoted as '$'"'"'b'
    return "'" + s.replace("'", "'\"'\"'") + "'"


def islice(iterable, *args):
    # islice('ABCDEFG', 2) --> A B
    # islice('ABCDEFG', 2, 4) --> C D
    # islice('ABCDEFG', 2, None) --> C D E F G
    # islice('ABCDEFG', 0, None, 2) --> A C E G
    s = slice(*args)
    start, stop, step = s.start or 0, s.stop or sys.maxsize, s.step or 1
    it = iter(range(start, stop, step))
    try:
        nexti = next(it)
    except StopIteration:
        # Consume *iterable* up to the *start* position.
        for i, element in zip(range(start), iterable):
            pass
        return
    try:
        for i, element in enumerate(iterable):
            if i == nexti:
                yield element
                nexti = next(it)
    except StopIteration:
        # Consume to *stop*.
        for i, element in zip(range(i + 1, stop), iterable):
            pass

def batched(iterable, n):
# batched('ABCDEFG', 3) --> ABC DEF G
    if n < 1:
        raise ValueError('n must be at least one')
    it = iter(iterable)
    while batch := tuple(islice(it, n)):
        yield batch


def parse_config(filepath:str) -> Dict[str, Any]:
    with open(filepath,"r") as f:
        config = yaml.safe_load(f)
    return config


def create_logger(app_name):
    """Create a logging interface"""
    logging_level = os.getenv('LOG_LVL', logging.INFO)
    logging.basicConfig(
        level=logging_level,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    logger = logging.getLogger(app_name)
    return logger


def exec_cmd(cmd: list[str], logger) -> str:
    cmd_str = quote(" ".join(list(map(lambda x: x.strip(), cmd))))
    logger.debug("[CMD] " + cmd_str)
    try:
        return subproc.check_output(cmd
               ,stderr=subproc.STDOUT).decode('utf-8').strip()
    except subproc.CalledProcessError as error:
        error.output = error.output.decode('utf-8').strip().replace('\n', '\n   | ')
        output=f"   | Command {cmd_str} failed with exit code {error.returncode}\n   | Output: {error.output}"
        raise subproc.CalledProcessError(error.returncode, error.cmd, output)


