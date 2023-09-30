import warnings
import matplotlib.pyplot as plt
import pickle
import numpy as np
import os
import pandas as pd

from sklearn.preprocessing import OneHotEncoder
from sklearn.compose import make_column_transformer
from sklearn.neighbors import KNeighborsClassifier
from sklearn.model_selection import GridSearchCV
from sklearn.exceptions import ConvergenceWarning
from sklearn.svm import SVC
from sklearn import metrics
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import train_test_split
from pandas.core.frame import DataFrame
from pandas.core.series import Series

RANDOM_STATE = 243
SHARED_DIR   = "/home/denilai/vbox_shared"
DUMP_DIR = os.path.join(os.path.dirname(__file__), "dumps")
RESULT_DIR   = os.path.join(os.path.dirname(__file__), "results")

coffee_tea_df = pd.read_csv(os.path.join(SHARED_DIR, "coffee_tea.csv"))
with open(os.path.join(DUMP_DIR, "coffee_tea_df_dump"), "wb") as f:
    pickle.dump(coffee_tea_df, f)


def confusion_matrix_plot(predicted, expected, filename):
    print(type(expected))
    print(type(predicted))
    confusion_matrix = metrics.confusion_matrix(predicted, expected)
    cm_display = metrics.ConfusionMatrixDisplay(confusion_matrix = confusion_matrix )
    cm_display.plot()
    plt.savefig(filename)
    plt.clf()

def knn(x_train, y_train,x_val, y_val):
    with warnings.catch_warnings():
        num_of_neighbors = np.arange(3,25,5)
        model_KNN = KNeighborsClassifier()
        parameters = {"n_neighbors": num_of_neighbors}
        grid_search = GridSearchCV(estimator = model_KNN, param_grid = parameters, cv = 6)
        grid_search.fit(x_train, y_train)
        print(f"Best scrore: {grid_search.best_score_}")
        print(f"Best estimator: {grid_search.best_estimator_}")
        knn_predicts = grid_search.predict(x_val)
        print(metrics.classification_report(knn_predicts, y_val))
        confusion_matrix_plot(knn_predicts, y_val,os.path.join(RESULT_DIR,"knn.png"))
    

def svc(x_train, y_train,x_val, y_val):
    with warnings.catch_warnings():
        warnings.simplefilter("ignore", category=UserWarning)
        param_kernel = ('linear', 'rbf', 'poly', 'sigmoid')
        parameters = {"kernel" : param_kernel}
        model = SVC()
        grid_search_svm = GridSearchCV(estimator = model, param_grid = parameters, cv = 6)
        grid_search_svm.fit(x_train,y_train)
        best_model = grid_search_svm.best_estimator_
        print(f"Best model: {best_model.kernel}")
        svm_predicts = best_model.predict(x_val)
        print(metrics.classification_report(svm_predicts, y_val))
        confusion_matrix_plot(svm_predicts, y_val,os.path.join(RESULT_DIR,"svc.png"))


def logistic_regression(x_train, y_train, x_val, y_val):
    with warnings.catch_warnings():
        warnings.simplefilter("ignore", category=ConvergenceWarning)
        model = LogisticRegression(random_state = 555)
        model.fit(x_train, y_train)
    y_predict = model.predict(x_val)
    confusion_matrix_plot(y_predict, y_val,os.path.join(RESULT_DIR,"logreg.png"))
    print(metrics.classification_report(y_predict, y_val))

if __name__ == "__main__" :
    # загрузка данных в дата фрейм
    df = coffee_tea_df
    print("Drop rows with NA values")
    df = df.dropna(axis = "index", how = "any")
    print(df.head())
    print(df.info())
    print(df.describe())
    
    y = df["Кофе или чай?"]
    X = df.drop(["Кофе или чай?", "Отметка времени"], axis = 1)
    print(X.isna().sum())
    print(y.isna().sum())

    # применение OneHotEncoder для подготовки данных. Для каждого уникального значения признака будет сформирован новый столбец (feature)"
    column_transformer = make_column_transformer((OneHotEncoder(), list(X.columns)), remainder='passthrough')
    X = pd.DataFrame(data = column_transformer.fit_transform(X), columns = column_transformer.get_feature_names_out())
    print(X.head())
    print(X.info())

    print("Cecking data...")
    print(X.isna().sum())
    print(X.describe())

    print("Create train and validation subset")
    x_train,x_val,y_train,y_val = train_test_split(X, y, train_size = 0.75,shuffle=True,random_state = RANDOM_STATE)


    print("x_train size:" , x_train.shape)
    print("y_train size:" , x_train.shape)
    print("x_val size:"   , y_val.shape)
    print("y_val size:"   , x_val.shape)
    print("Logistic regression")
    logistic_regression(x_train,y_train,x_val, y_val)
    print("SVC")
    svc(x_train,y_train,x_val, y_val)
    print("KNN")
    knn(x_train, y_train, x_val, y_val)
