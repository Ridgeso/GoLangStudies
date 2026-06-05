#include <iostream>
#include <vector>
#include <thread>
#include <atomic>

std::atomic<long long> total(0);

void worker(int id)
{
    long long sum = 0;
    for (int i = 0; i < 10000; i++)
    {
        sum += i * id;
    }
    total.fetch_add(sum, std::memory_order_relaxed);
}

int main()
{
    const int numWorkers = 1000;
    std::vector<std::thread> threads;
    threads.reserve(numWorkers);
    for (int i = 0; i < numWorkers; i++)
    {
        threads.emplace_back(worker, i);
    }
    for (auto& t : threads)
    {
        t.join();
    }
    printf("%lld\n", total.load());
}
