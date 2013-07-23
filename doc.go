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
// N. Ferguson and B. Schneier.  Fortuna is a cryptographically strong
// pseudo-random number generator, typical use cases include
// generation of keys in cryptographic ciphers and session tokens for
// web apps.
//
// The fortuna random number generator consists of two parts: The
// accumulator is the high-level random number generator which can use
// caller-provided randomness (i.e. timings between the user's key
// presses) to generate truely random output.  The accumulator uses a
// pseudo random number generator to generate its output; this
// generator is also available as a stand-alone component.
//
// Accumulator
//
// The class Accumulator provides the usual way to use the Fortuna
// random number generator.  A new Accumulator can be allocated
// using the NewAccumulator() function:
//
//     acc, err := fortuna.NewAccumulator(aes.NewCipher, seedFileName)
//     if err != nil {
//         panic("cannot initialise the RNG: " + err.Error())
//     }
//     defer acc.WriteSeedFile(seedFileName)
//
// The argument seedFileName is the name of a small file where
// randomness is stored between runs of the program.  The program must
// be able to both read and write this file, and the contents must be
// kept confidential.  The file is updated every 10 minutes during the
// program run and should also be updated on shutdown using a call to
// acc.WriteSeedFile(seedFileName).
//
// If the seedFileName argument equals the empty string "", no seed
// file is used.  In this case, the generator must be seeded before
// random output can be generated.  The easiest way to initialise the
// generator in this case is to call acc.SetInitialSeed().
//
// After the generatator is initialised, randomness can be extracted
// using the RandomData() method:
//
//     data := acc.RandomData(16)
//
// Finally, the program using the Accumulator should continuously
// collect randomness from the environment and submit this randomness
// to the Accumulator for incorporation into the random output.
// For example, code like the following could be used to submit
// the inter-request times in a web-server to the Accumulator:
//
//     source := uint8(100)
//     pool := uint8(0)
//     lastRequest := time.Now()
//     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//         now := time.Now()
//         dt := now.Sub(lastRequest)
//         lastRequest = now
//         acc.AddRandomEvent(source, pool, []byte(dt.String()))
//         pool = (pool + 1) % 32
//
//         ...
//     })
//
// Generator
//
// The call Generator provides a pseudo random number generator which
// forms the basis of the Accumulator described above.  New instances
// of the Fortuna pseudo random number generator can be created using
// the NewGenerator().  The function newCipher should normally be
// aes.NewCipher from the crypto/aes package, but the Serpent or
// Twofish ciphers can also be used:
//
//     gen := fortuna.NewGenerator(aes.NewCipher)
//
// Before use, the generator must be seeded using the Seed() or
// Reseed() functions:
//
//     gen.Seed(1234)
//
// Uniformly distributed random bytes can the be extracted using the
// PseudoRandomData() method:
//
//     data := gen.PseudoRandomData(16)
//
// Generator implements the rand.Source interface and thus the
// functions from the math/rand package can be used to obtain pseudo
// random samples from more complicated distributions.
package fortuna
