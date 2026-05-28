def sieve(n):
    is_prime = bytearray([1]) * (n + 1)
    is_prime[0] = is_prime[1] = 0
    i = 2
    while i * i <= n:
        if is_prime[i]:
            is_prime[i*i::i] = bytearray(len(is_prime[i*i::i]))
        i += 1
    print(sum(is_prime))

sieve(5_000_000)
