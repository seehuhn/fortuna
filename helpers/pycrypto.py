#! /usr/bin/env python
#
# Generate reference data for "generator_test.go"

from sys import stdout
from Crypto.Random.Fortuna.FortunaGenerator import AESGenerator

rng = AESGenerator()

stdout.write("\tcorrect := []byte{")
rng.reseed("\1\2\3\4")
for i, x in enumerate(rng.pseudo_random_data(100)):
    if i % 15 == 0:
        stdout.write("\n\t\t")
    else:
        stdout.write(" ")
    stdout.write("%d,"%ord(x))
stdout.write("\n\t}\n")

stdout.write("\tcorrect = []byte{")
rng.reseed("\5")
for i, x in enumerate(rng.pseudo_random_data(100)):
    if i % 15 == 0:
        stdout.write("\n\t\t")
    else:
        stdout.write(" ")
    stdout.write("%d,"%ord(x))
stdout.write("\n\t}\n")
