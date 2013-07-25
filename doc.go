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
// The Fortuna random number generator consists of two parts: The
// accumulator collects caller-provided randomness (i.e. timings
// between the user's key presses).  This randomness is then used to
// seed a pseudo random number generator.  During operation, the
// randomness from the accumulator is also used to periodically reseed
// the generator, thus allowing to recover from limited compromises of
// the generator's state.  The accumulator and the generator are
// described in separate sections, below.
//
// Accumulator
//
// The usual way to use the Fortuna random number generator is by
// creating an object of type Accumulator.  A new Accumulator can be
// allocated using the NewAccumulator() function:
//
//     acc, err := fortuna.NewAccumulator(aes.NewCipher, seedFileName)
//     if err != nil {
//         panic("cannot initialise the RNG: " + err.Error())
//     }
//     defer acc.Close()
//
// The argument seedFileName is the name of a file where a small
// amount of randomness can be stored between runs of the program.
// The program must be able to both read and write this file, and the
// contents must be kept confidential.  While the accumulator is in
// use, the file is updated every 10 minutes.  If a seed file is used,
// the Accumulator should be closed using the Close() method after
// use.
//
// If the seedFileName argument equals the empty string "", no seed
// file is used.  In this case, the generator must be seeded before
// random output can be generated.  The easiest way to initialise the
// generator in this case is to call acc.SetInitialSeed().
//
// After the generatator is initialised, randomness can be extracted
// using the RandomData() and Read() methods.  For example, a slice of
// 16 random bytes can be obtained using the following command.
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
//     seq := uint(0)
//     lastRequest := time.Now()
//     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//         now := time.Now()
//         dt := now.Sub(lastRequest)
//         lastRequest = now
//         acc.AddRandomEvent(source, seq, []byte(dt.String()))
//         seq += 1
//
//         ...
//     })
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
// Before use, the generator must be seeded using the Seed() or
// Reseed() functions:
//
//     gen.Seed(1234)
//
// Uniformly distributed random bytes can then be extracted using the
// PseudoRandomData() method:
//
//     data := gen.PseudoRandomData(16)
//
// Generator implements the rand.Source interface and thus the
// functions from the math/rand package can be used to obtain pseudo
// random samples from more complicated distributions.
package fortuna
