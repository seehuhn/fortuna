#! /usr/bin/env python
#
# Use the "Python Cryptography Toolkit" from
# https://www.dlitz.net/software/pycrypto/ to generate reference data
# for "generator_test.go"

import sys
import time

from Crypto.Random.Fortuna.FortunaGenerator import AESGenerator
from Crypto.Random.Fortuna.FortunaAccumulator import FortunaAccumulator

######################################################################
# part 1: test data for AESGenerator

rng = AESGenerator()

rng.reseed("\1\2\3\4")
sys.stdout.write("\tcorrect := []byte{")
for i, x in enumerate(rng.pseudo_random_data(100)):
    if i % 15 == 0:
        sys.stdout.write("\n\t\t")
    else:
        sys.stdout.write(" ")
    sys.stdout.write("%d,"%ord(x))
sys.stdout.write("\n\t}\n")

sys.stdout.write("\tcorrect = []byte{")
for i, x in enumerate(rng.pseudo_random_data((1<<20) + 100)[-100:]):
    if i % 15 == 0:
        sys.stdout.write("\n\t\t")
    else:
        sys.stdout.write(" ")
    sys.stdout.write("%d,"%ord(x))
sys.stdout.write("\n\t}\n")

rng.reseed("\5")
sys.stdout.write("\tcorrect = []byte{")
for i, x in enumerate(rng.pseudo_random_data(100)):
    if i % 15 == 0:
        sys.stdout.write("\n\t\t")
    else:
        sys.stdout.write(" ")
    sys.stdout.write("%d,"%ord(x))
sys.stdout.write("\n\t}\n\n\n")

######################################################################
# part 2: test data for FortunaAccumulator

acc = FortunaAccumulator()

acc.add_random_event(0, 0, "\0"*32)
acc.add_random_event(0, 0, "\0"*32)
for i in range(1000):
    acc.add_random_event(1, i%32, "\1\2")
sys.stdout.write("\tcorrect := []byte{")
for i, x in enumerate(acc.random_data(100)):
    if i % 15 == 0:
        sys.stdout.write("\n\t\t")
    else:
        sys.stdout.write(" ")
    sys.stdout.write("%d,"%ord(x))
sys.stdout.write("\n\t}\n")

acc.add_random_event(0, 0, "\0"*32)
acc.add_random_event(0, 0, "\0"*32)
sys.stdout.write("\tcorrect = []byte{")
for i, x in enumerate(acc.random_data(100)):
    if i % 15 == 0:
        sys.stdout.write("\n\t\t")
    else:
        sys.stdout.write(" ")
    sys.stdout.write("%d,"%ord(x))
sys.stdout.write("\n\t}\n")

time.sleep(0.2)

sys.stdout.write("\tcorrect = []byte{")
for i, x in enumerate(acc.random_data(100)):
    if i % 15 == 0:
        sys.stdout.write("\n\t\t")
    else:
        sys.stdout.write(" ")
    sys.stdout.write("%d,"%ord(x))
sys.stdout.write("\n\t}\n")
