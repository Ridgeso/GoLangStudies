#include <iostream>
#include <vector>

using Mat = std::vector<std::vector<double>>;

int main() {
    const int n = 400;
    Mat a(n, std::vector<double>(n));
    Mat b(n, std::vector<double>(n));
    Mat c(n, std::vector<double>(n, 0.0));

    for (int i = 0; i < n; i++)
    {
        for (int j = 0; j < n; j++)
        {
            a[i][j] = (i+j) * 0.01;
            b[i][j] = (i-j) * 0.01;
        }
    }

    for (int i = 0; i < n; i++)
    {
        for (int k = 0; k < n; k++)
        {
            double aik = a[i][k];
            for (int j = 0; j < n; j++)
            {
                c[i][j] += aik * b[k][j];
            }
        }
    }

    printf("%.6f\n", c[n-1][n-1]);
}
