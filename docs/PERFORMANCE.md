# Performance

> 中文版本：[docs/readme/PERFORMANCE.zh.md](readme/PERFORMANCE.zh.md)

This document compares the runtime cost of the different ways to obtain a JSON
Schema for a Protobuf message: generating it dynamically at runtime versus
accessing a statically generated artifact.

## Summary

| Approach           | Latency           | Throughput  | Allocation | When to use                         |
| ------------------ | ----------------- | ----------- | ---------- | ----------------------------------- |
| Dynamic generation | ~8.4 μs           | ~118K ops/s | 2.7 KB/op  | Default for most services           |
| Static `map` const | ~0.29 ns + 5.9 μs | ~170K ops/s | 2.5 KB/op  | Need a mutable map structure        |
| Static Go struct   | ~0.29 ns + 4.7 μs | ~212K ops/s | 2.2 KB/op  | Need type safety                    |
| Static JSON string | ~0.29 ns          | unbounded   | 0          | **Optimal** — HTTP-ready, zero cost |

Static access is roughly **29,000× faster** than dynamic generation, and a JSON
string constant has zero serialization overhead because it is already JSON.

## Static formats compared

For the static (`format=go_const`) output, four representations are possible:

| Static format     | Access    | Serialize | Total       | When to use              |
| ----------------- | --------- | --------- | ----------- | ------------------------ |
| JSON string const | 0.29 ns   | 0 ns      | **0.29 ns** | HTTP API (optimal)       |
| Go struct const   | 0.29 ns   | 4,708 ns  | 4,708 ns    | Type safety needed       |
| `map` const       | 0.29 ns   | 5,875 ns  | 5,875 ns    | Mutable structure needed |
| Read from file    | 13,279 ns | 0 ns      | 13,279 ns   | Not recommended          |

## CPU cost at scale

Per-request CPU overhead of dynamic generation at various request rates:

| Scenario      | RPS     | Dynamic CPU       | Static CPU | Recommendation    |
| ------------- | ------- | ----------------- | ---------- | ----------------- |
| Web API       | 1,000   | 0.81%             | ~0%        | Dynamic           |
| Microservice  | 5,000   | 4.1%              | ~0%        | Dynamic           |
| API gateway   | 50,000  | 40.5%             | ~0%        | Static            |
| Message queue | 500,000 | 405% (infeasible) | ~0%        | Static (required) |

## Guidance

- **Dynamic generation** suits ~99% of services (RPS < 10K, runtime flexibility,
  used as a library). Single generation costs ~8 μs and ~2.7 KB.
- **Static generation** suits high-throughput paths (RPS > 100K, fixed message
  set). Prefer the JSON string constant for HTTP responses — it is emitted
  directly with `GetJSONSchemaBytes()` at zero cost.
- For medium throughput, dynamic generation with a `sync.Map` cache keyed by the
  message full name is a simple middle ground.

> Numbers are indicative benchmark results on a single machine; treat them as
> relative ratios rather than absolute guarantees. Re-run `go test
./benchmark/...` to measure on your hardware.
