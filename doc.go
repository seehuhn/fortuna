// The Fortuna random number generator by N. Ferguson and B. Schneier
// Copyright (C) 2013  Jochen Voss <voss@seehuhn.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package fortuna implements the Fortuna random number generator by
// Niels Ferguson and Bruce Schneier.  Fortuna is a cryptographically
// strong pseudo-random number generator; typical use cases include
// generation of keys in cryptographic ciphers and session tokens for
// web apps.
//
// The homepage of this package is at
// http://www.seehuhn.de/pages/fortuna .  Please send any comments or
// bug reports to the program's author, Jochen Voss <voss@seehuhn.de>.
//
// The Fortuna random number generator consists of two parts: The
// accumulator collects caller-provided randomness (e.g. timings
// between the user's key presses).  This randomness is then used to
// seed a pseudo random number generator.  During operation, the
// randomness from the accumulator is also used to periodically reseed
// the generator, thus allowing to recover from limited compromises of
// the generator's state.  The accumulator and the generator are
// described in separate sections, below.
//
//
// Accumulator
//
// The usual way to use the Fortuna random number generator is by
// creating an object of type Accumulator.  A new Accumulator can be
// allocated using the NewRNG() function:
//
//     rng, err := fortuna.NewRNG(seedFileName)
//     if err != nil {
//         panic("cannot initialise the RNG: " + err.Error())
//     }
//     defer rng.Close()
//
// The argument seedFileName is the name of a file where a small
// amount of randomness can be stored between runs of the program.
// The program must be able to both read and write this file, and the
// contents must be kept confidential.  If the seedFileName argument
// equals the empty string "", no entropy is stored between runs.  In
// this case, the initial seed is only based on the current time of
// day, the current user name, the list of currently installed network
// interfaces, and output of the system random number generator.  Not
// using a seed file can lead to more predictable output in the
// initial period after the generator has been created; a seed file
// must be used in security sensitive applications.
//
// If a seed file is used, the Accumulator must be closed using the
// Close() method after use.
//
// Randomness can be extracted from the Accumulator using the
// RandomData() and Read() methods.  For example, a slice of 16 random
// bytes can be obtained using the following command:
//
//     data := rng.RandomData(16)
//
//
// Entropy Pools
//
// The Accumulator uses 32 entropy pools to collect randomness from
// the environment.  The use of external entropy helps to recover from
// situations where an attacker obtained (partial) knowledge of the
// generator state.
//
// Any program using the Fortuna generator should continuously collect
// random/unpredictable data and should submit this data to the
// Accumulator.  For example, code like the following could be used to
// submit the times between requests in a web-server:
//
//     sink := rng.NewEntropyTimeStampSink()
//     defer close(sink)
//     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//         sink <- time.Now()
//
//         ...
//     })
//
//
// Generator
//
// The Generator class provides a pseudo random number generator which
// forms the basis of the Accumulator described above.  New instances
// of the Fortuna pseudo random number generator can be created using
// the NewGenerator() function.  The argument newCipher should
// normally be aes.NewCipher from the crypto/aes package, but the
// Serpent or Twofish ciphers can also be used:
//
//     gen := fortuna.NewGenerator(aes.NewCipher)
//
// The generator can be seeded using the Seed() or Reseed() methods:
//
//     gen.Seed(1234)
//
// The method .Seed() should be used if reproducible output is
// required, whereas .Reseed() can be used to add entropy in order to
// achieve less predictable output.
//
// Uniformly distributed random bytes can then be extracted using the
// .PseudoRandomData() method:
//
//     data := gen.PseudoRandomData(16)
//
// Generator implements the rand.Source interface and thus the
// functions from the math/rand package can be used to obtain pseudo
// random samples from more complicated distributions.
package fortuna
