#include <iostream>

int64_t fibRec(int64_t n) {
    if (n <= 1)
    {
        return n;
    }
    return fibRec(n-1) + fibRec(n-2);
}

int main()
{
    printf("%lld\n", fibRec(40));
    return 0;
}