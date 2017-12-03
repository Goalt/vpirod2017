import sys
import csv

sys.path.insert(0, "./mincemeatpy")
import mincemeat

if len(sys.argv) != 2:
    print("Wrong number of parameters")

data = []

inputFile = open(sys.argv[1], "r")
csvReader = csv.reader(inputFile)
iterCsvReader = iter(csvReader)
next(iterCsvReader)
for line in iterCsvReader:
    names = line[2].replace(" and ", ", ").split(", ")
    for n in names:
        data.append((n,line[3]))

datasource = dict(enumerate(data))

def mapfn(k, v):
    import nltk
    tokenizer = nltk.tokenize.RegexpTokenizer(r"\w+")
    speech = tokenizer.tokenize(v[1].lower())
    yield v[0], speech

def reducefn(k, vs):
    result = set()
    for i in vs:
        result = result | set(i)
    return len(result)

s = mincemeat.Server()
s.datasource = datasource
s.mapfn = mapfn
s.reducefn = reducefn

results = s.run_server(password="changeme")

outputFile = open("results.csv", "w")
outputFile.write("Name,Language Wealth\r\n")
for i in results.keys():
    outputFile.write(i + "," + str(results[i]) + "\r\n")