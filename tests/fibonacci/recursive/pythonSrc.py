import sys
sys.setrecursionlimit(100_000)

def fib_rec(n):
    if n <= 1:
        return n
    return fib_rec(n-1) + fib_rec(n-2)

print(fib_rec(40))
