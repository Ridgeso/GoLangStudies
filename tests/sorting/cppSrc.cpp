#include <iostream>
#include <vector>
#include <algorithm>
using namespace std;

int main() {
    const int n = 10000000;
    vector<int> arr(n);
    for (int i = 0; i < n; i++)
    {
        arr[i] = (i * 1000003 + 7) % n;
    }
    sort(arr.begin(), arr.end());
    printf("%d %d\n", arr[0], arr[n-1]);
}
