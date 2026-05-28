"""
Massive multi-language benchmark runner.
Compiles and executes Go, C++, and Python implementations of the same
algorithms, measures compile time + execution time, and writes JSON results.

Usage:
    python3 benchmark.py                  # run all benchmarks, 3 runs each
    python3 benchmark.py --runs 5         # override run count
    python3 benchmark.py --filter sort    # only benchmarks matching name
    python3 benchmark.py --no-go          # skip Go
    python3 benchmark.py --no-cpp         # skip C++
    python3 benchmark.py --no-python      # skip Python
    python3 benchmark.py --timeout 120    # per-run timeout in seconds
"""

import argparse
import json
import os
import shutil
import stat
import subprocess
import sys
import time
import tempfile
from dataclasses import dataclass, field, asdict
from pathlib import Path
from typing import Optional


ROOT       = Path(__file__).parent
GO_SRC     = ROOT / "go_src"
CPP_SRC    = ROOT / "cpp_src"
PY_SRC     = ROOT / "py_src"
RESULTS    = ROOT / "results"
BUILD_DIR  = ROOT / "build"


@dataclass
class RunResult:
    exit_code: int
    stdout: str
    stderr: str
    wall_seconds: float


@dataclass
class BenchmarkResult:
    name: str
    language: str
    compile_s: Optional[float]
    runs_s: list[float]        = field(default_factory=list)
    exec_mean_s: float         = 0.0
    exec_min_s: float          = 0.0
    exec_max_s: float          = 0.0
    total_s: float             = 0.0   # compile + exec_mean
    output: str                = ""
    error: str                 = ""
    skipped: bool              = False
    skip_reason: str           = ""


GO_BIN = shutil.which("go") or "/usr/lib/go/bin/go"


def _run(cmd: list[str], timeout: int, cwd: Optional[Path] = None) -> RunResult:
    t0 = time.perf_counter()
    try:
        p = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=timeout,
            cwd=str(cwd) if cwd else None,
        )
        elapsed = time.perf_counter() - t0
        return RunResult(p.returncode, p.stdout.strip(), p.stderr.strip(), elapsed)
    except subprocess.TimeoutExpired:
        elapsed = time.perf_counter() - t0
        return RunResult(-1, "", f"TIMEOUT after {timeout}s", elapsed)
    except Exception as exc:
        elapsed = time.perf_counter() - t0
        return RunResult(-2, "", str(exc), elapsed)


def _compile_go(src: Path, out: Path, timeout: int) -> tuple[bool, float, str]:
    r = _run([GO_BIN, "build", "-o", str(out), str(src)], timeout)
    return r.exit_code == 0, r.wall_seconds, r.stderr


def _compile_cpp(src: Path, out: Path, timeout: int) -> tuple[bool, float, str]:
    r = _run(["g++", "-O2", "-std=c++17", "-pthread", "-o", str(out), str(src)], timeout)
    return r.exit_code == 0, r.wall_seconds, r.stderr


def run_benchmark(
    name: str,
    language: str,
    src: Path,
    n_runs: int,
    timeout: int,
) -> BenchmarkResult:
    res = BenchmarkResult(name=name, language=language, compile_s=None)

    binary: Optional[Path] = None

    if language == "go":
        if not src.exists():
            res.skipped, res.skip_reason = True, f"source not found: {src}"
            return res
        binary = BUILD_DIR / f"{name}_go"
        ok, ctime, cerr = _compile_go(src, binary, timeout)
        res.compile_s = ctime
        if not ok:
            res.skipped, res.skip_reason = True, f"compile error:\n{cerr}"
            res.error = cerr
            return res

    elif language == "cpp":
        if not src.exists():
            res.skipped, res.skip_reason = True, f"source not found: {src}"
            return res
        binary = BUILD_DIR / f"{name}_cpp"
        ok, ctime, cerr = _compile_cpp(src, binary, timeout)
        res.compile_s = ctime
        if not ok:
            res.skipped, res.skip_reason = True, f"compile error:\n{cerr}"
            res.error = cerr
            return res

    elif language == "python":
        if not src.exists():
            res.skipped, res.skip_reason = True, f"source not found: {src}"
            return res
        res.compile_s = 0.0   # interpreted

    for run_idx in range(n_runs):
        if language in ("go", "cpp"):
            cmd = [str(binary)]
        else:
            cmd = [sys.executable, str(src)]

        r = _run(cmd, timeout)

        if r.exit_code not in (0,):
            res.skipped = True
            res.skip_reason = f"run {run_idx+1} exited {r.exit_code}: {r.stderr}"
            res.error = r.stderr
            return res

        res.runs_s.append(r.wall_seconds)
        if run_idx == 0:
            res.output = r.stdout   # capture first run output

    if res.runs_s:
        res.exec_mean_s = sum(res.runs_s) / len(res.runs_s)
        res.exec_min_s  = min(res.runs_s)
        res.exec_max_s  = max(res.runs_s)
    res.total_s = (res.compile_s or 0.0) + res.exec_mean_s
    return res


BENCHMARKS = [
    ("fibonacci_iter", "fibonacci.go",      "fibonacci.cpp",      "fibonacci.py",      "cpu",         "Fibonacci(40) iterative"),
    ("fibonacci_rec",  "fib_recursive.go",  "fib_recursive.cpp",  "fib_recursive.py",  "cpu",         "Fibonacci(42) recursive — exponential call tree"),
    ("prime_sieve",    "prime_sieve.go",    "prime_sieve.cpp",    "prime_sieve.py",    "cpu",         "Sieve of Eratosthenes up to 5 million"),
    ("matrix_mul",     "matrix_mul.go",     "matrix_mul.cpp",     "matrix_mul.py",     "cpu",         "400x400 matrix multiplication (naive triple-loop)"),
    ("sort_ints",      "sort_ints.go",      "sort_ints.cpp",      "sort_ints.py",      "cpu",         "Sort 10 million integers (stdlib sort)"),
    ("json_roundtrip", "json_roundtrip.go", "json_roundtrip.cpp", "json_roundtrip.py", "data",        "Encode + decode 200 000 objects as JSON"),
    ("string_build",   "string_build.go",   "string_build.cpp",   "string_build.py",   "data",        "Build string from 500 000 formatted tokens"),
    ("hashmap",        "hashmap.go",        "hashmap.cpp",        "hashmap.py",        "data",        "Insert + read 2 million key-value pairs"),
    ("file_io",        "file_io.go",        "file_io.cpp",        "file_io.py",        "io",          "Write then read 1 million lines to/from disk"),
    ("concurrency",    "concurrency.go",    "concurrency.cpp",    "concurrency.py",    "concurrency", "1 000 goroutines/threads doing arithmetic"),
]


RESET  = "\033[0m"
BOLD   = "\033[1m"
GREEN  = "\033[32m"
YELLOW = "\033[33m"
RED    = "\033[31m"
CYAN   = "\033[36m"
DIM    = "\033[2m"


def _color_time(t: float) -> str:
    if t < 0.1:
        return f"{GREEN}{t:.4f}s{RESET}"
    if t < 1.0:
        return f"{YELLOW}{t:.4f}s{RESET}"
    return f"{RED}{t:.4f}s{RESET}"


def print_result(r: BenchmarkResult):
    lang_pad = r.language.ljust(7)
    if r.skipped:
        print(f"  {DIM}{lang_pad}  SKIPPED — {r.skip_reason[:80]}{RESET}")
        return
    compile_str = f"compile={_color_time(r.compile_s)}" if r.compile_s is not None and r.language != "python" else f"compile={DIM}n/a{RESET}    "
    print(f"  {CYAN}{lang_pad}{RESET}  {compile_str}  "
          f"exec_mean={_color_time(r.exec_mean_s)}  "
          f"exec_min={_color_time(r.exec_min_s)}  "
          f"total={_color_time(r.total_s)}")


def main():
    parser = argparse.ArgumentParser(description="Multi-language benchmark runner")
    parser.add_argument("--runs",      type=int, default=3,   help="exec runs per benchmark (default 3)")
    parser.add_argument("--timeout",   type=int, default=180, help="per-run timeout seconds (default 180)")
    parser.add_argument("--filter",    type=str, default="",  help="only run benchmarks whose name contains this string")
    parser.add_argument("--no-go",     action="store_true",   help="skip Go")
    parser.add_argument("--no-cpp",    action="store_true",   help="skip C++")
    parser.add_argument("--no-python", action="store_true",   help="skip Python")
    parser.add_argument("--output",    type=str, default="",  help="path to write JSON results (default: results/run_<timestamp>.json)")
    args = parser.parse_args()

    BUILD_DIR.mkdir(parents=True, exist_ok=True)
    RESULTS.mkdir(parents=True, exist_ok=True)

    languages = []
    if not args.no_go:
        languages.append(("go",     GO_SRC))
    if not args.no_cpp:
        languages.append(("cpp",    CPP_SRC))
    if not args.no_python:
        languages.append(("python", PY_SRC))

    benchmarks = BENCHMARKS
    if args.filter:
        benchmarks = [b for b in benchmarks if args.filter.lower() in b[0].lower()]

    print(f"\n{BOLD}═══════════════════════════════════════════════════════{RESET}")
    print(f"{BOLD}  Multi-language Benchmark Suite{RESET}")
    print(f"{BOLD}═══════════════════════════════════════════════════════{RESET}")
    print(f"  Languages : {', '.join(l for l,_ in languages)}")
    print(f"  Benchmarks: {len(benchmarks)}")
    print(f"  Runs each : {args.runs}")
    print(f"  Timeout   : {args.timeout}s per run")
    print()

    all_results: list[BenchmarkResult] = []
    total_start = time.perf_counter()

    for (bname, go_f, cpp_f, py_f, category, desc) in benchmarks:
        src_map = {"go": GO_SRC / go_f, "cpp": CPP_SRC / cpp_f, "python": PY_SRC / py_f}

        print(f"{BOLD}[{category.upper()}] {bname}{RESET}  {DIM}{desc}{RESET}")

        for (lang, src_dir) in languages:
            src = src_map[lang]
            r = run_benchmark(bname, lang, src, args.runs, args.timeout)
            all_results.append(r)
            print_result(r)

        print()

    elapsed_total = time.perf_counter() - total_start

    print(f"\n{BOLD}═══════════════════════════════════════════════════════{RESET}")
    print(f"{BOLD}  Summary — execution time (mean), normalized to Python{RESET}")
    print(f"{BOLD}═══════════════════════════════════════════════════════{RESET}")

    header = f"  {'Benchmark':<22} {'Python':>10} {'Go':>10} {'C++':>10}  {'Go speedup':>12}  {'C++ speedup':>12}"
    print(header)
    print("  " + "─" * (len(header) - 2))

    for (bname, *_) in benchmarks:
        row: dict[str, BenchmarkResult] = {
            r.language: r for r in all_results if r.name == bname
        }
        py_t  = row.get("python")
        go_t  = row.get("go")
        cpp_t = row.get("cpp")

        def fmt(r: Optional[BenchmarkResult]) -> str:
            if r is None or r.skipped: return "  skip"
            return f"{r.exec_mean_s:>8.3f}s"

        def speedup(base: Optional[BenchmarkResult], target: Optional[BenchmarkResult]) -> str:
            if base is None or target is None:
                return "     n/a"
            if base.skipped or target.skipped:
                return "     n/a"
            if target.exec_mean_s == 0:
                return "     ∞"
            ratio = base.exec_mean_s / target.exec_mean_s
            return f"{ratio:>8.2f}x"

        print(f"  {bname:<22} {fmt(py_t)} {fmt(go_t)} {fmt(cpp_t)}  "
              f"{speedup(py_t, go_t):>12}  {speedup(py_t, cpp_t):>12}")

    print(f"\n  Total wall time: {elapsed_total:.1f}s")

    ts = int(time.time())
    out_path = Path(args.output) if args.output else RESULTS / f"run_{ts}.json"

    payload = {
        "timestamp": ts,
        "runs_per_benchmark": args.runs,
        "timeout_s": args.timeout,
        "total_wall_s": round(elapsed_total, 3),
        "results": [asdict(r) for r in all_results],
    }
    out_path.write_text(json.dumps(payload, indent=2))
    print(f"\n  Results written → {out_path}\n")


if __name__ == "__main__":
    main()
