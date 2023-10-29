import networkx as nx
import csv

import os
import sys

utils_path = os.path.join(os.path.dirname(os.getcwd()),"pract_2")
sys.path.append(utils_path)

from utils import create_logger, read_from_сsv

logger = create_logger("Graph")


def main():
    # Указание id одногруппников, для которых будут найдены центральности
    classmates =  [824691883, 785314237, 802674640, 276648630]
    G = nx.Graph()
    # Загрузка данных из csv-файла
    pairs = read_from_сsv("friends.csv")
    G.add_edges_from(pairs)
    # Нахождение центральности по близоcти
    logger.info(f"Find closeness centrality for {classmates}")
    clos_centrality = list(map(lambda x: nx.closeness_centrality(G, u=str(x)), classmates))
    print(clos_centrality)
    # Нахождение центральности по посредничеству
    logger.info(f"Find betweenness centrality")
    betwen_centrality = nx.betweenness_centrality(G)
    print(betwen_centrality)
    # Нахождение центральности по собственному значению
    logger.info(f"Find eigenvector centrality")


if __name__ == "__main__":
    main()


def closeness_centrality(G, nodes, *args, **kwargs):
    logger.info(f"Find cliseness centrality for {nodes}")
    res = list(map(lambda x: nx.closeness_centrality(G, u=x), nodes))
    return res
    
