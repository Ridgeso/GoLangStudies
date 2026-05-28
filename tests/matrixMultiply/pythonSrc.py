n = 400
a = [[(i+j)*0.01 for j in range(n)] for i in range(n)]
b = [[(i-j)*0.01 for j in range(n)] for i in range(n)]
c = [[0.0]*n for _ in range(n)]

for i in range(n):
    for k in range(n):
        aik = a[i][k]
        for j in range(n):
            c[i][j] += aik * b[k][j]

print(f"{c[n-1][n-1]:.6f}")
