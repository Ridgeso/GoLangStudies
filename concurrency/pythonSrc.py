import threading

total = 0
lock = threading.Lock()

def worker(wid):
    global total
    s = sum(i * wid for i in range(10_000))
    with lock:
        total += s

threads = [threading.Thread(target=worker, args=(i,)) for i in range(1000)]
for t in threads: t.start()
for t in threads: t.join()
print(total)
