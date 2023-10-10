from utils import parse_config
import psycopg2
import logging


logging.basicConfig(level = logging.INFO)
logger = logging.getLogger("vk")

config = parse_config("./config.yaml")
try:
  pg = config["postgres"]
except KeyError as e:
    raise SystemExit("Wrong cofnig. Section `postgres` is expected") from e



def create_friends_with_table(conn):
  query = """
    create table if not exists friends_with (
      id1 text,
      id2 text
    )
  """
  with conn:
    with conn.cursor() as cur:
      cur.execute(query)
      logging.info("Table `FRIENDS_WITH` is created (if not exists)")
  conn.close()

def copy_table_from_f(conn, f, table:str, **kwargs):
  with conn:
    with conn.cursor() as cur:
      cur.copy_from(f, table, **kwargs)
  conn.close()



conn = psycopg2.connect(host = pg["host"],
                        dbname = pg["database"],
                        user = pg["user"],
                        password = pg["password"])


if __name__== "__main__":
  with open("friends.csv", "r") as f:
    copy_table_from_f(conn, f, "friends_with", sep=",")






