#include <iostream>
#include <unordered_map>

int main() {
    const int N = 2000000;
    std::unordered_map<int,int> m;
    m.reserve(N * 2);
    for (int i = 0; i < N; i++)
    {
        m[i] = i * 2;
    }
    int64_t sum = 0;
    for (int i = 0; i < N; i++)
    {
        sum += m[i];
    }
    printf("%lld\n", sum);
}