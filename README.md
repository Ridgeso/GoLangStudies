# Go vs C++ vs Python Benchmark Suite

Compare performance across 10 algorithms in 3 languages.

## Quick start

```bash
python3 benchmark.py
```

Requires `go`, `g++`, and `python3` on your PATH.

## Options

```bash
python3 benchmark.py --runs 5          # more runs = better averages (default: 3)
python3 benchmark.py --filter sort     # run only benchmarks matching a name
python3 benchmark.py --no-python       # skip a language
python3 benchmark.py --timeout 120     # per-run timeout in seconds
python3 benchmark.py --output out.json # custom results path
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
├── benchmark.py         # runner — the only script you need
├── tests/               # sources of test examples
    ├── concurrency/     #
    ├── fibonacci/       #
    ├── fibRecursive/    #
    ├── fileIO/          #
    ├── hashmap/         #
    ├── jsonParser/      #
    ├── matrixMultiply/  #
    ├── primeSieve/      #
    ├── sorting/         #
    ├── stringBuilder/   #
├── build/               # compiled binaries (auto-created)
└── results/             # JSON output per run (auto-created)
```

## Results

Each run saves a JSON file to `results/`. Times reported: compile, execution (mean/min/max), and total.

## Key findings (from real run)

- Matrix multiply: Go ~99× faster than Python, C++ ~184×
- Concurrency: Go goroutines ~57× faster than Python threads (GIL)
- File I/O and sorting: much closer, Go ~3–5×
- Go compiles in ~0.15s per file; C++ ~0.30–0.44s
