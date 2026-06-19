#include <iostream>
#include <vector>

int32_t sieve(int32_t n) {
    auto isPrime = std::vector<bool>(n+1, true); //std::vector<char> would actually be faster
    
    for (int32_t i = 2; i*i <= n; i++)
    {
        if (isPrime[i])
        {
            for (int32_t j = i*i; j <= n; j += i)
            {
                isPrime[j] = false;
            }
        }
    }
    
    int32_t count = 0;
    for (int32_t i = 2; i <= n; i++)
    {
        if (isPrime[i]) count++;
    }
    return count;
}

int32_t main()
{
    printf("%d\n", sieve(5000000));
    return 0;
}
