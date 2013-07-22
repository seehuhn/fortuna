#! /usr/bin/env python
#
# Use the "Python Cryptography Toolkit" from
# https://www.dlitz.net/software/pycrypto/ to generate reference data
# for "generator_test.go"

from sys import stdout
from Crypto.Random.Fortuna.FortunaGenerator import AESGenerator

rng = AESGenerator()

rng.reseed("\1\2\3\4")
stdout.write("\tcorrect := []byte{")
for i, x in enumerate(rng.pseudo_random_data(100)):
    if i % 15 == 0:
        stdout.write("\n\t\t")
    else:
        stdout.write(" ")
    stdout.write("%d,"%ord(x))
stdout.write("\n\t}\n")

stdout.write("\tcorrect = []byte{")
for i, x in enumerate(rng.pseudo_random_data((1<<20) + 100)[-100:]):
    if i % 15 == 0:
        stdout.write("\n\t\t")
    else:
        stdout.write(" ")
    stdout.write("%d,"%ord(x))
stdout.write("\n\t}\n")

rng.reseed("\5")
stdout.write("\tcorrect = []byte{")
for i, x in enumerate(rng.pseudo_random_data(100)):
    if i % 15 == 0:
        stdout.write("\n\t\t")
    else:
        stdout.write(" ")
    stdout.write("%d,"%ord(x))
stdout.write("\n\t}\n")
