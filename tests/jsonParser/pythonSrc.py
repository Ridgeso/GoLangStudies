import json

N = 200_000
items = [
    {"id": i, "name": f"item_{i}", "value": i * 1.23456, "tags": ["alpha","beta","gamma"]}
    for i in range(N)
]
encoded = json.dumps(items)
decoded = json.loads(encoded)
print(len(decoded))
