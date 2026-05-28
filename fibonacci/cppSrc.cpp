#include <iostream>

uint64_t fib(uint64_t n)
{
    if (n <= 1)
    {
        return n;
    }
    uint64_t a = 0;
    uint64_t b = 1;
    for (uint64_t i = 2; i <= n; i++)
    {
        uint64_t t = a + b;
        a = b;
        b = t;
    }
    return b;
}

int32_t main()
{
    printf("%llu\n", fib(40));
    return 0;
}
