## ⚙️ Test Environment

| Parameter        | Value                     |
|------------------|---------------------------|
| CPU Cores        | 8                         |
| Benchmark Host   | `127.0.0.1` (localhost)   |
| Multi-threaded DB| `1234`                    |
| Redis Port       | `6379`                    |

---

## 🔹 Benchmark 1: **SET Command**

### Command
```bash
./redis/src/redis-benchmark -p 1234 -t set -n 1000000 -r 1000000 --threads 4
```

### Results

| System | Throughput (req/s) | Avg Latency (ms) | P50 | P95 | P99 | Max |
|--------|--------------------|------------------|-----|-----|-----|-----|
| **Multi-threaded In-Memory DB** | **45,191.61** | 1.006 | 0.983 | 1.335 | 1.943 | 70.911 |
| **Redis (Single-threaded)**     | **46,442.50** | 1.025 | 0.983 | 1.583 | 2.055 | 77.567 |

✅ **Observation:**  
The performance of the multi-threaded database is **comparable to Redis** for write operations, maintaining **low latency** and stable throughput.

---

## 🔹 Benchmark 2: **GET Command**

### Command
```bash
./redis/src/redis-benchmark -n 1000000 -t get -c 500 -h 127.0.0.1 -p 1234 -r 1000000 --threads 4
```

### Results

| System | Throughput (req/s) | Avg Latency (ms) | P50 | P95 | P99 | Max |
|--------|--------------------|------------------|-----|-----|-----|-----|
| **Multi-threaded In-Memory DB** | **57,009.29** | 8.626 | 8.319 | 11.471 | 12.295 | 29.647 |
| **Redis (Single-threaded)**     | **43,853.88** | 11.253 | 11.063 | 15.239 | 17.071 | 37.023 |

✅ **Observation:**  
The multi-threaded design achieves **~30% higher throughput** and **lower latency** under high concurrency (`-c 500`).  
This indicates strong **parallel read scalability**.

---

## 🔹 Benchmark 3: **ZADD Command**

### Command
```bash
./redis/src/redis-benchmark   -n 1000000   -c 200   -r 100000   -p 1234   --threads 4   "ZADD" "zset:__rand_int__" "__rand_int__" "member:__rand_int__"
```

### Results

| System | Throughput (req/s) | Avg Latency (ms) | P50 | P95 | P99 | Max |
|--------|--------------------|------------------|-----|-----|-----|-----|
| **Multi-threaded In-Memory DB** | **64,383.21** | 3.003 | 3.007 | 3.959 | 5.167 | 33.919 |
| **Redis (Single-threaded)**     | **51,203.27** | 3.799 | 3.639 | 5.895 | 7.119 | 12.495 |

✅ **Observation:**  
The multi-threaded system delivers **~25% higher throughput** and **better latency distribution** for sorted set operations (`ZADD`), which are CPU-intensive.

---

## 🔹 Benchmark 4: **ZRANGE Command**

### Command
```bash
./redis/src/redis-benchmark   -n 1000000   -c 500   -r 100000   -p 1234   --threads 4   "ZRANGE" "myzset:__rand_int__" "0" "-1"
```

### Results

| System | Throughput (req/s) | Avg Latency (ms) | P50 | P95 | P99 | Max |
|--------|--------------------|------------------|-----|-----|-----|-----|
| **Multi-threaded In-Memory DB** | **51,135.20** | 9.564 | 9.183 | 12.719 | 14.431 | 119.039 |
| **Redis (Single-threaded)**     | **44,835.00** | 10.949 | 10.719 | 14.879 | 17.167 | 42.303 |


---

## 📊 Summary Comparison

| Command | Metric | Multi-threaded DB | Redis | Improvement |
|----------|---------|------------------|--------|--------------|
| **SET** | Throughput | 45,191 req/s | 46,442 req/s | ≈ Same |
| **GET** | Throughput | 57,009 req/s | 43,854 req/s | **+30%** |
| **ZADD** | Throughput | 64,383 req/s | 51,203 req/s | **+25%** |
| **ZRANGE** | Throughput | 51,135 req/s | 44,835 req/s | **+15%** |

---

---

## 🔹 Test 1: **Fan-out (Create Post)**

### Description
- 125 users concurrently call `CreatePost()`.
- Each user has **1,000 followers**.
- Measure total processing time from `CreatePost()` start to completion of `CachePost()` for all followers.

### Results

| Metric | Multi-threaded DB | Redis |
|---------|-------------------------------|-------------------|
| Total users (concurrent) | 125 | 125 |
| Total elapsed time | **5.183s** | **7.294s** |
| Average latency | **4.086s** | **7.193s** |
| P95 latency | **5.178s** | **7.292s** |
| P99 latency | **5.182s** | **7.293s** |

✅ **Observation:**  
The multi-threaded in-memory DB completed the fan-out process **~30% faster** than Redis, showing better scalability under concurrent write-heavy load.

---

## 🔹 Test 2: **GenerateNewsfeed**

### Description
- 10,000 users concurrently open the homepage (trigger `GenerateNewsfeed(userId)` simultaneously).
- Each user has a feed of **200 posts** stored in a Redis ZSET-like structure.
- Measure throughput and latency.

### Results

| Metric | Multi-threaded DB | Redis |
|---------|-------------------------------|-------------------|
| Requests | 125 | 125 |
| Total time | **4.411s** | **9.184s** |
| Throughput | **28.34 req/s** | **13.61 req/s** |
| Average latency | **2.858s** | **8.464s** |
| P95 latency | **4.399s** | **9.180s** |
| P99 latency | **4.408s** | **9.182s** |

✅ **Observation:**  
The multi-threaded DB handled **GenerateNewsfeed** nearly **2× faster** than Redis, maintaining lower latency even under high concurrency.

---

## ⚡ Summary

| Test | Metric | Multi-threaded DB | Redis | Improvement |
|------|---------|------------------|--------|--------------|
| Fan-out (Create Post) | Total Time | 5.18s | 7.29s | **+30% faster** |
| GenerateNewsfeed | Throughput | 28.34 req/s | 13.61 req/s | **+2.08× faster** |

---

## 💻 CPU Utilization (Illustration)

Below is an ASCII visualization of **CPU utilization** between Redis and the Multi-threaded system.

```
+-----------------------------------------------------------+
|                    CPU Utilization                        |
+--------------------+--------------------------------------+
| Redis (Single-thread)      | ██████████  (1 core @100%)  |
| Multi-threaded Database    | ████ ████ ████ ████  (8 cores active) |
+--------------------+--------------------------------------+
```

🟥 **Redis:** Utilizes a single CPU core effectively (single-threaded event loop).  
⬛ **Multi-threaded DB:** Distributes workload across multiple cores, achieving better parallel throughput.

---
