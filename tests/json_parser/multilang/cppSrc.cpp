#include <iostream>
#include <string>
#include <vector>
#include <sstream>

struct Item
{
    int id;
    std::string name;
    double value;
};

std::string encode(const std::vector<Item>& items)
{
    std::string out = "[";
    for (size_t i = 0; i < items.size(); i++)
    {
        if (i) out += ",";
        std::ostringstream ss;
        ss << "{\"id\":" << items[i].id
           << ",\"name\":\"" << items[i].name << "\""
           << ",\"value\":" << items[i].value << "}";
        out += ss.str();
    }
    return out + "]";
}

int countObjects(const std::string& s)
{
    int c = 0;
    for (char ch : s) if (ch == '{') c++;
    return c;
}

int main()
{
    const int N = 200000;
    std::vector<Item> items(N);
    for (int i = 0; i < N; i++)
        items[i] = {i, "item_" + std::to_string(i), i * 1.23456};
    printf("%d\n", countObjects(encode(items)));
}
