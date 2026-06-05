n = 10_000_000
arr = [(i * 1_000_003 + 7) % n for i in range(n)]
arr.sort()
print(arr[0], arr[-1])
