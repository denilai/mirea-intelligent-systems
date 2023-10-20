
from neo4j import GraphDatabase
from typing import List, Any, Dict, Tuple, Optional, Set
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

    def close(self):
      self.driver.close()


        

    def _load_person_from_csv(self, tx, filepath):
        src = f"file:///{filepath}"
        print(filepath)
        query  = "load csv from $src as line\n"
        query += "merge (p1:Person {id:line[0]})\n"
        query += "merge (p2:Person {id:line[1]})\n"
        query += "merge (p1)-[:IS_FRIENDS_WITH]->(p2);"
        logger.info(f"Load {filepath}")
        tx.run(query, src = src)


    def _add_person_friends(self, tx, root_id:str, leaf_ids:List[str]) -> None:
        query = f"""merge (root:Person {{id:$root_id}}) """
        params = {"root_id":root_id}
        for i, leaf_id  in enumerate(leaf_ids):
            query += f"""
            merge (leaf_{i}:Person {{id:$leaf_{i}id}})
            merge (root)-[:IS_FRIENDS_WITH]->(leaf_{i})
            """
            params[f"leaf_{i}id"]=leaf_id
        tx.run(query, parameters = params)

    def _add_person(self, tx, person_id):
        result = tx.run("CREATE (p: Person {id: $person_id})", person_id = person_id)

    def _detach_delete_persons(self, tx):
        logger.info("Detach delete all Person nodes")
        result = tx.run("MATCH (n:Person) DETACH DELETE n;")

    def detach_delete_persons(self):
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._detach_delete_persons)


    def add_person_friends(self, root_id, leaf_ids:List[str]) -> None:
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._add_person_friends,root_id, leaf_ids)

    def load_person_from_csv(self, filepath):
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._load_person_from_csv, filepath)

    def add_person(self, person_id):
        with self.driver.session(database = self.database) as session:
            session.execute_write(self._add_person, person_id)
            print(f"Person with id = {person_id} was created") 