import sys
import mincemeat
import csv

if len(sys.argv) != 6:
    print("Wrong number of parameters")

sizeA = (sys.argv[2], sys.argv[3])
sizeB = (sys.argv[4], sys.argv[5])

inputFile = open(sys.argv[1], 'r')
csvReader = csv.reader(inputFile)
csvReader = iter(csvReader)
next(csvReader)

datasource = {}
for line in csvReader:
    if line[0] == "a":
        datasource[("a", int(line[1]), int(line[2]), int(sizeB[1]))] = float(line[3])
    else:
        datasource[("b", int(line[1]), int(line[2]), int(sizeA[0]))] = float(line[3])
inputFile.close()

def mapfn(k, v):
    if k[0] == "a":
        for i in range(k[3]):
            yield (k[1], i), (k[0], k[2], v)
    else:
        for i in range(k[3]):
            yield (i, k[1]), (k[0], k[2], v)


def reducefn(k, vs):
    A = []
    B = []

    for i in vs:
        if i[0] == "a":
            A.append(i)
        else:
            B.append(i)

    A = sorted(A, key=lambda item: item[1])
    B = sorted(B, key=lambda item: item[1])
    cnt = 0
    result = 0
    for i in range(len(A)):
        for j in range(cnt, len(B)):
            if A[i][1] == B[j][1]:
                cnt = j + 1
                result = (result + (A[i][2] * B[j][2]) % 97) % 97
                break

    return result

s = mincemeat.Server()
s.datasource = datasource
s.mapfn = mapfn
s.reducefn = reducefn
results = s.run_server(password="changeme")

outputFile = open("results.csv", 'w')
outputFile.write("matrix,i,j,value")
for k in results.keys():  # out
    if results[k] != 0:
        outputFile.write("\r\n" + "c," + str(k[0]) + ',' + str(k[1]) + ',' + str(results[k]))

outputFile.close()
