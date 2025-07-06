# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Go implementation of the Fortuna cryptographically strong pseudorandom number generator by Ferguson and Schneier. It's a single-module Go package that provides both high-level (`Accumulator`) and low-level (`Generator`) interfaces for secure random number generation.

## Development Commands

### Testing
```bash
go test          # Run all tests
go test -v       # Run tests with verbose output
go test -run TestName  # Run specific test
```

### Code Quality
```bash
go vet           # Check for common Go mistakes
go fmt .         # Format all Go files
go build         # Build the package
```

### Documentation
```bash
godoc -http=:6060  # Start local documentation server
go doc             # Show package documentation
```

## Architecture

### Core Components

1. **Accumulator** (`accumulator.go`): High-level interface that combines entropy collection with random number generation
   - Manages 32 entropy pools for collecting environmental randomness
   - Handles seed file persistence for entropy between program runs
   - Thread-safe for concurrent access
   - Main entry point: `NewRNG(seedFileName)`

2. **Generator** (`generator.go`): Low-level cryptographic PRNG
   - Uses AES in counter mode (configurable cipher)
   - Implements `rand.Source` and `rand.Source64` interfaces
   - Automatic rekeying after generating 2^16 blocks
   - Key size: 32 bytes (SHA256 hash size)

3. **Entropy Collection** (`entropy.go`): Environmental randomness gathering
   - Provides channels for submitting entropy data
   - `NewEntropyDataSink()` for arbitrary data
   - `NewEntropyTimeStampSink()` for timing-based entropy
   - Distributes entropy across pools using round-robin

4. **Seed Management** (`seed.go`): Persistent entropy storage
   - File-based seed storage with file locking
   - Automatic seed file updates every 10 minutes
   - Initial seeding from system sources when no seed file exists

### Key Constants
- `numPools = 32`: Number of entropy pools
- `minPoolSize = 32`: Minimum bytes per pool before reseeding
- `minReseedInterval = 100ms`: Minimum time between reseeds
- `maxBlocks = 65536`: Maximum blocks before rekeying
- `keySize = 32`: Internal key size in bytes

## Usage Patterns

### Basic Usage
```go
rng, err := fortuna.NewRNG("seed.file")
if err != nil {
    // handle error
}
defer rng.Close()

data := rng.RandomData(16)  // Get 16 random bytes
```

### Entropy Collection
```go
sink := rng.NewEntropyTimeStampSink()
defer close(sink)
// Submit timing events
sink <- time.Now()
```

### Low-level Generator
```go
gen := fortuna.NewGenerator(aes.NewCipher)
gen.Seed(1234)
data := gen.PseudoRandomData(16)
```

## Testing

All core functionality has comprehensive tests:
- `accumulator_test.go`: Accumulator functionality, seed file handling
- `generator_test.go`: PRNG output, seeding, compatibility
- `entropy_test.go`: Entropy pool management
- `seed_test.go`: Seed file operations
- `common_test.go`: Utility functions
- `flocksyscall_test.go`: File locking

## Dependencies

- `github.com/seehuhn/sha256d`: Double SHA-256 hashing
- Standard library: `crypto/aes`, `crypto/cipher`, `crypto/rand`, `crypto/sha256`

## Platform Support

- Cross-platform Go code
- File locking implementation varies by platform (`flocksyscall.go` vs `flockstub.go`)
- Minimum Go version: 1.13