def main():
    N = 100_000
    queries = 2_000

    data = [i * 2 for i in range(N)]

    hits = 0
    for i in range(queries):
        key = (i * 7) % (N * 2)
        
        for v in data:
            if v == key:
                hits += 1
                break

    print(hits)

if __name__ == "__main__":
    main()
