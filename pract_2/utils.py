from typing import Dict, Any
import yaml
import logging
import os


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
