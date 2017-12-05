import os
import sys
import mincemeat

if len(sys.argv) != 2:
    print("Wrong number of parameters")

data = []

files = os.listdir("./" + sys.argv[1])
files = sorted(files)

for f in files:
    inputFile = open("./" + sys.argv[1] + "/" + f, "r")

    line = inputFile.read()
    data.append((f, line))

    inputFile.close()

datasource = dict(enumerate(data))

# print(datasource)

def mapfn(k, v):
    import nltk
    tokenizer = nltk.tokenize.RegexpTokenizer(r"\w+")
    speech = tokenizer.tokenize(v[1].lower())
    for sp in speech:
        yield (v[0], sp), 1

def reducefn(k, vs):
    return sum(vs)

def mapfn2(k, v):
    yield k[1], (k[0], v) 

def reducefn2(k, vs):
    result = {}
    for i in vs:
        result[i[0]] = i[1]
    return result

s = mincemeat.Server()
s.datasource = datasource
s.mapfn = mapfn
s.reducefn = reducefn
results = s.run_server(password="changeme")

s.datasource = results
s.mapfn = mapfn2
s.reducefn = reducefn2
results = s.run_server(password="changeme")

outputFile = open("results.csv", "w")
outputFile.write("Word")
for f in files:
    outputFile.write("," + f)
outputFile.write("\r\n")

for i in results.keys():
    outputFile.write(i)
    for f in files:
        if results[i].get(f) is None:
            outputFile.write("," + "0")
        else:
            outputFile.write("," + str(results[i].get(f)))
    outputFile.write("\r\n")

outputFile.close()