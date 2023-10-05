
import csv
import random
import math


# TODO: функция будет работать только с ограниченным набором строковых данных. Нет валидации и нормализации значений.
def work_with_data(data):
  for i in range(len(data)):
    for j in range(len(data[i])):
      if data[i][j] == 'Да' or data[i][j] == 'М' or data[i][j] == 'Сова' or data[i][j] == 'Чай':
        data[i][j] = 1
      elif data[i][j] == 'Нет' or data[i][j] == 'Ж' or data[i][j] == 'Жаворонок'or data[i][j] == 'Кофе':
        data[i][j] = 0
      elif data[i][j] == '':
        data[i][j] = 0
      else:
        match data[i][j]:
          case '5-6 утра':
            data[i][j] = 0
          case '6-7 утра':
            data[i][j] = 0.33
          case '7-8 утра':
            data[i][j] = 0.66
          case 'Позднее 8 утра':
            data[i][j] = 1
          case 'Менее 3 часов':
            data[i][j] = 0
          case 'От 3 до 6 часов':
            data[i][j] = 0.33
          case '7-8 часов':
            data[i][j] = 0.66
          case 'Более 8 часов':
            data[i][j] = 1
          case 'Менее 3 часов':
            data[i][j] = 0
          case 'От 3 до 6 часов':
            data[i][j] = 0.33
          case '7-8 часов':
            data[i][j] = 0.66
          case 'Более 8 часов':
            data[i][j] = 1
          case 'Подмосковье':
            data[i][j] = 0
          case 'Восточный административный округ':
            data[i][j] = 0.125
          case 'Северо-Восточный административный округ':
            data[i][j] = 0.25
          case 'Северный административный округ':
            data[i][j] = 0.375
          case 'Северо-западный административный округ':
            data[i][j] = 0.5
          case 'Юго-Западный административный округ':
            data[i][j] = 0.625
          case 'Южный административный округ':
            data[i][j] = 0.75
          case 'Юго-Восточный административный округ':
            data[i][j] = 0.875
          case 'Центральный административный округ':
            data[i][j] = 1
  return data

# TODO: нет проверки на корректность аргументов
def split_for_train_and_test(data,percent):
  test = list(data)
  train = []
  while (len(test)/(len(train)+1)>percent):
    toss = random.randint(0, len(test)-1)
    train.append(test[toss])
    test.pop(toss)
  return(train, test)


# TODO: удалить
def sortirovka(neighbors):
  return []

# TODO: нет проверки параметров. 
def find_neighbors(train, test_i,k):
    all_points = list(train)
    point = list(test_i)
    neighbors = []
    for i in range(len(all_points)):
      polusum = []
      for j in range(len(all_points[i])-1):
        polusum.append(((all_points[i][j+1] - point[j+1])**2))
      summa = 0.0
      for z in range(len(polusum)):
        summa = summa + polusum[z]
      neighbors.append([math.sqrt(summa),all_points[i][0]])
    tea_count = 0
    neighbors = sorted(neighbors)
    for l in range(len(neighbors)):
      if (l < k-1):
        tea_count += neighbors[l][0]
      elif (l == k-1) and ((tea_count + neighbors[l][1])*2 != k):
        tea_count += neighbors[l][0]
    if ((tea_count*2) < k):
      prediction = 0
    else:
      prediction = 1
    return prediction



# TODO: опечатка в слове accuracy
def get_accurtacy(predictions, test):
  all = 0
  for i in range(len(test)):
    if (predictions[i] == test[i][0]):
      all = 1 + all
  accuracy = all / len(test)
  return accuracy


data = []
with open("test_result.csv", ) as file_name:
    file_read = csv.reader(file_name)
    data = list(file_read)
data = work_with_data(data)

print ("original:")
print (data)
print (len(data))
train, test = split_for_train_and_test(data, 0.3)

for K in range(3,10):
  predictions = []
  for i in range(len(test)):
    predictions.append(find_neighbors(train,test[i],K))
  accuracy = get_accurtacy(predictions,test)
  print(F'K = {K} Точность предсказания = {accuracy}')
  print (predictions)

print ("train:")
print (len(train))
print ("test:")
print (len(test))
print ("original:")
print (len(data))
