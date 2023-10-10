from typing import Dict, Any
import yaml
def parse_config(filepath:str) -> Dict[str, Any]:
    with open(filepath,"r") as f:
        config = yaml.safe_load(f)
    return config
