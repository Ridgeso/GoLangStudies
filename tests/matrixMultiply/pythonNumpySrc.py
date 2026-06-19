import numpy as np

n = 400
a = np.fromfunction(lambda i, j: (i + j) * 0.01, (n, n))
b = np.fromfunction(lambda i, j: (i - j) * 0.01, (n, n))
c = a @ b

print(f"{c[n-1][n-1]:.6f}")