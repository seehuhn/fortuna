Fortuna
=======

An implementation of Ferguson and Schneier's Fortuna_ random number
generator in Go.

Copyright (C) 2013  Jochen Voss

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

The homepage of this package is at <http://www.seehuhn.de/pages/fortuna>.
Please send any comments or bug reports to the program's author,
Jochen Voss <voss@seehuhn.de>.

.. _Fortuna: http://en.wikipedia.org/wiki/Fortuna_(PRNG)

Overview
--------

Fortuna is a `cryptographically strong`_ random number generator (RNG).
The term "cryptographically strong" indicates that even a very clever
and active attacker, who knows some of the random outputs of the RNG,
cannot use this knowledge to predict future or past outputs.  This
property allows, for example, to use the output of the RNG to generate
keys for encryption schemes, and to generate session tokens for web
pages.

.. _cryptographically strong: http://en.wikipedia.org/wiki/Cryptographically_secure_pseudorandom_number_generator

Random number generators are hard to implement and easy to get wrong;
even seemingly small details can make a huge difference to the
security of the method.  For this reason, this implementation tries to
follow the original description of the Fortuna generator (chapter 10
of [FS03]_) as closely as possible.  In addition, some effort was made
to ensure that, given idential seeds, the output of this
implementation coincides with the output of the implementation from
the `Python Cryptography Toolkit`_.

.. [FS03] Niels Ferguson, Bruce Schneier: *Practical Cryptography*, Wiley, 2003.
.. _Python Cryptography Toolkit: https://www.dlitz.net/software/pycrypto/

Installation
------------

This package can be installed using the ``go get`` command::

    go get github.com/seehuhn/fortuna

Usage
-----

The fortuna random number generator consists of two parts: The
accumulator collects caller-provided randomness (i.e. timings between
the user's key presses).  This randomness is then used to seed a
pseudo random number generator.  During operation, the randomness from
the accumulator is also used to periodically reseed the generator,
thus allowing to recover from limited compromises of the generator's
state.  Both, the accumulator and the generator are described in
separte sections, below.

Accumulator
...........

The class ``Accumulator`` provides the usual way to use the Fortuna
random number generator.  A new ``Accumulator`` can be allocated
using the ``NewAccumulator()`` function::

    acc, err := fortuna.NewAccumulator(aes.NewCipher, seedFileName)
    if err != nil {
	panic("cannot initialise the RNG: " + err.Error())
    }
    defer acc.Close()

The argument ``seedFileName`` is the name of a file where a small
amount of randomness can be stored between runs of the program.  The
program must be able to both read and write this file, and the
contents must be kept confidential.  While the accumulator is in use,
the file is updated every 10 minutes.  If a seed file is in used, the
Accumulator should be closed using the ``Close()`` method after use.

If the ``seedFileName argument`` equals the empty string ``""``, no
seed file is used.  In this case, the generator must be seeded before
random output can be generated.  The easiest way to initialise the
generator in this case is to call ``acc.SetInitialSeed()``.

After the generatator is initialised, randomness can be extracted
using the ``RandomData()`` and ``Read()`` methods.  For example, a
slice of 16 random bytes can be obtained using the following command::

    data := acc.RandomData(16)

Finally, the program using the Accumulator should continuously collect
randomness from the environment and submit this randomness to the
Accumulator for incorporation into the random output.  For example,
code like the following could be used to submit the inter-request
times in a web-server to the Accumulator::

    source := uint8(100)
    pool := uint8(0)
    lastRequest := time.Now()
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	dt := now.Sub(lastRequest)
	lastRequest = now
	acc.AddRandomEvent(source, pool, []byte(dt.String()))
	pool = (pool + 1) % 32

	...
    })

Generator
.........

The ``Generator`` class provides a pseudo random number generator
which forms the basis of the accumulator described above.  New
instances of the Fortuna pseudo random number generator can be created
using the ``NewGenerator()`` function.  The argument ``newCipher``
should normally be ``aes.NewCipher`` from the ``crypto/aes`` package,
but the Serpent_ or Twofish_ ciphers can also be used::

    gen := fortuna.NewGenerator(aes.NewCipher)

.. _Serpent: http://en.wikipedia.org/wiki/Serpent_(cipher)
.. _Twofish: http://en.wikipedia.org/wiki/Twofish

Before use, the generator must be seeded using the ``Seed()`` or
``Reseed()`` functions::

    gen.Seed(1234)

Uniformly distributed random bytes can then be extracted using the
``PseudoRandomData()`` method::

    data := gen.PseudoRandomData(16)

``Generator`` implements the ``rand.Source`` interface and thus the
functions from the ``math/rand`` package can be used to obtain pseudo
random samples from more complicated distributions.

Detailed usage instructions are available via the package's online
help, either on godoc.org_ or on the command line::

    go doc github.com/seehuhn/fortuna

.. _godoc.org: http://godoc.org/github.com/seehuhn/fortuna
