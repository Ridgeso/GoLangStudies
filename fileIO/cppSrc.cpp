#include <iostream>
#include <fstream>
#include <string>

int main()
{
    const char* path = "tmp/bench_io_cpp.txt";

    std::ofstream out(path);
    for (int i = 0; i < 1000000; i++)
    {
        out << i << "\n";
    }
    out.close();

    std::ifstream in(path);
    std::string line;
    int count = 0;
    while (std::getline(in, line))
    {
        count++;
    }
    in.close();
    std::remove(path);
    std::cout << count << "\n";
}
