import os
from pathlib import Path

tmp = Path("tmp")
path = tmp / "bench_io_py.txt"

tmp.mkdir()

with open(path, "w+") as f:
    for i in range(1_000_000):
        f.write(f"{i}\n")

with open(path) as f:
    count = sum(1 for _ in f)

os.remove(path)
tmp.rmdir()
print(count)
