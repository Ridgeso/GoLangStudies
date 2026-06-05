#include <iostream>
#include <string>

int32_t main()
{
    auto result = std::string{""};
    result.reserve(8'000'000);
    for (int32_t i = 0; i < 500'000; i++)
    {
        result += "word";
        result += std::to_string(i);
        result += ' ';
    }
    printf("%d\n", static_cast<int32_t>(result.size()));
}
