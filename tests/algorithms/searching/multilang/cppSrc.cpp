#include <iostream>
#include <vector>

int main() {
    const int N = 100000;
    const int queries = 2000;

    std::vector<int> data(N);
    for (int i = 0; i < N; ++i) {
        data[i] = i * 2;
    }

    int hits = 0;
    for (int i = 0; i < queries; ++i) {
        int key = (i * 7) % (N * 2);
        
        for (int v : data) {
            if (v == key) {
                hits++;
                break;
            }
        }
    }

    printf("%d\n", hits);
    return 0;
}
