N = 2_000_000
m = {i: i * 2 for i in range(N)}
total = sum(m[i] for i in range(N))
print(total)
