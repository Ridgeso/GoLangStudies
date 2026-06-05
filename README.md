# Go vs C++ vs Python Benchmark Suite

Compare performance across 10 algorithms in 3 languages.

## Quick start

```bash
go run benchmark.go
```

Requires `go`, `g++`, and `python3` on your PATH.

## Options

```bash
go run benchmark.go --runs 5               # more runs = better averages (default: 3)
go run benchmark.go --filter sort          # run only benchmarks matching a name
go run benchmark.go --no-python            # skip a language
go run benchmark.go --timeout 120          # per-run timeout in seconds
go run benchmark.go --output out.json      # custom results path
go run benchmark.go --cpp-compiler=clang++ # change c++ compiler to custom
```

## Benchmarks

| Name | What it tests |
|---|---|
| `fibonacci_iter` | Iterative Fibonacci(40) |
| `fibonacci_rec` | Recursive Fibonacci(42) — exponential call tree |
| `prime_sieve` | Sieve of Eratosthenes up to 5 million |
| `matrix_mul` | 400×400 matrix multiplication |
| `sort_ints` | Sort 10 million integers |
| `json_roundtrip` | Encode + decode 200,000 objects |
| `string_build` | Build string from 500,000 tokens |
| `hashmap` | Insert + read 2 million key-value pairs |
| `file_io` | Write then read 1 million lines |
| `concurrency` | 1,000 goroutines / threads |

## Structure

```
benchmarks/
┌── benchmark.go         # runner — the only script you need
└── tests/               # sources of test examples
    ├── concurrency/     #
    ├── fibonacci/       #
    ├── fibRecursive/    #
    ├── fileIO/          #
    ├── hashmap/         #
    ├── jsonParser/      #
    ├── matrixMultiply/  #
    ├── primeSieve/      #
    ├── sorting/         #
    ├── stringBuilding/  #
┌── build/               # compiled binaries (auto-created)
└── results/             # JSON output per run (auto-created)
```

## Results

Each run saves a JSON file to `results/`. Times reported: compile, execution (mean/min/max), and total.
