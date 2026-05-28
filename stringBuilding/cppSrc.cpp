#include <iostream>
#include <string>
    
int main()
{
    std::string result;
    result.reserve(8000000);
    for (int i = 0; i < 500000; i++)
    {
        result += "word";
        result += std::to_string(i);
        result += ' ';
    }
    printf("%d\n", static_cast<int>(result.size()));
}
