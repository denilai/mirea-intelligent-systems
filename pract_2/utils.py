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


