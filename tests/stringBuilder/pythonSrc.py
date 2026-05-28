parts = []
for i in range(500_000):
    parts.append(f"word{i} ")
result = "".join(parts)
print(len(result))
