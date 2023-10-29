
from neo4j import GraphDatabase
from typing import Any, Dict, Tuple, Optional, Set
from utils import create_logger

logger = create_logger("Neo4j")


class Neo4jAgent:
    def __init__(self,
                 host:str,
                 port:str,
                 user:str,
                 password:str,
                 database:str
    ):
      endpoint = f"bolt://{host}:{port}"
      self.driver = GraphDatabase.driver(endpoint, auth=(user,password))
      self.database = database


    def add_node_leafs(self, root_label: str, leaf_label: str, relationship: str, root_id:int, leaf_ids:list[int]) -> None:
        def _add_node_leafs(tx, root_label:str, leaf_label:str, relationship:str, root_id:int, leaf_ids:list[int]) -> None:
            query = f"merge (root: {root_label} {{id: $root_id}})"
            params = {"root_id":root_id}
            for i, leaf_id  in enumerate(leaf_ids):
                query += f"""
                merge (leaf_{i}: {leaf_label} {{id: $leaf_{i}id}})
                merge (root)-[:{relationship}]->(leaf_{i})
                """
                params[f"leaf_{i}id"]=leaf_id
            tx.run(query, parameters = params)

        with self.driver.session(database = self.database) as session:
            session.execute_write(_add_node_leafs, root_label, leaf_label, relationship, root_id, leaf_ids)


    def detach_delete_nodes(self, label:str):
        def _detach_delete_nodes(tx, label:str):
            logger.info(f"Detach delete all `{label}` nodes")
            result = tx.run(f"MATCH (n:{label}) DETACH DELETE n;")
        with self.driver.session(database = self.database) as session:
            session.execute_write(_detach_delete_nodes, label)


    def load_from_csv(self, filepath, label:str, relationship:str):
        def _load_from_csv(tx, filepath:str, label:str, relationship:str):
            src = f"file:///{filepath}"
            print(filepath)
            query  = "load csv from $src as line\n"
            query += f"merge (p1:{label} {{id:line[0]}})\n"
            query += f"merge (p2:{label} {{id:line[1]}})\n"
            query += f"merge (p1)-[:{relationship}]->(p2);"
            logger.info(f"Load {filepath}")
            tx.run(query, src = src)

        with self.driver.session(database = self.database) as session:
            session.execute_write(_load_from_csv, filepath,label, relationship)

    def add_node(self, person_id):
        def _add_node(tx, label:str, node_id:int):
            tx.run(f"CREATE (p:{label} {{id: $node_id}})", node_id = node_id)

        with self.driver.session(database = self.database) as session:
            session.execute_write(_add_node, person_id)
            print(f"Person with id = {person_id} was created") 


    def close(self):
      self.driver.close()

