# KVCache BYOE (Real Engine) Sample

[Back to KVCache Index](../README.md) | English | [中文](./README_zh.md)

## Layout

- `pixiu/conf.yaml`: BYOE Pixiu template (env-driven)
- `request.sh`: manual request script
- `test/pixiu_test.go`: smoke test (requires external BYOE env)
- `verify.sh`: verification script for metrics-oriented validation

## Prerequisites

- reachable `VLLM_ENDPOINT` with `/tokenize`
- reachable `LMCACHE_ENDPOINT` with `/lookup` `/pin` `/compress` `/evict`
- Pixiu binary or Pixiu source repo

## Quick Start (CLI)

1. Configure env and run verification:

```bash
cd ai/kvcache/real-engine
export VLLM_ENDPOINT="http://<vllm-host>:<port>"
export LMCACHE_ENDPOINT="http://<lmcache-host>:<port>"
./verify.sh
```

2. Send manual request:

```bash
./request.sh
```

3. Run test case:

```bash
go test -v ./test/pixiu_test.go
```

## Acceptance Metrics

- lookup success rate
- preferred endpoint hit rate
- p95 latency comparison (without cache vs with cache)
