import os
import sys
import random

if len(sys.argv) != 7:
    print("Wrong number of parameters")

outputFile = open(sys.argv[1], 'w')
i = int(sys.argv[2])
j = int(sys.argv[3])
k = int(sys.argv[4])
densityA = float(sys.argv[5])
densityB = float(sys.argv[6])

random.seed()

cntNumbers = int(i * j * densityA)
numbersSetA = set()
for l in range(cntNumbers):
	(x,y) = (random.randint(0, i-1), random.randint(0, j-1))
	while (x,y) in numbersSetA:
		(x,y) = (random.randint(0, i-1), random.randint(0, j-1))
	(x,y,val) = (x,y,random.uniform(-100, 100))
	numbersSetA.add((x,y,val))

cntNumbers = int(j * k * densityB)
numbersSetB = set()
for l in range(cntNumbers):
	(x,y) = (random.randint(0, j-1), random.randint(0, k-1))
	while (x,y) in numbersSetB:
		(x,y) = (random.randint(0, j-1), random.randint(0, k-1))
	(x,y,val) = (x,y,random.uniform(-100, 100))
	numbersSetB.add((x,y,val))

outputFile.write("matrix,i,j,value")
for i in numbersSetA:
	outputFile.write("\r\n" + "a," + str(i[0]) + ',' + str(i[1]) + ',' + str(i[2]))
for i in numbersSetB:
	outputFile.write("\r\n" + "b," + str(i[0]) + ',' + str(i[1]) + ',' + str(i[2]))	
