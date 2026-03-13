# KVCache Mock Sample

[Back to KVCache Index](../README.md) | English | [中文](./README_zh.md)

## Layout

- `pixiu/conf.yaml`: Pixiu configuration
- `server/controller/main.go`: mock LMCache controller (`/lookup` `/pin` `/compress` `/evict`)
- `server/engine-a/main.go`: mock engine A (`/tokenize` + `/v1/chat/completions`)
- `server/engine-b/main.go`: mock engine B (`/v1/chat/completions`)
- `request.sh`: request script
- `test/pixiu_test.go`: integration test
- `run.sh`: one-command startup + validation

## Quick Start (CLI)

1. Start sample services and Pixiu:

```bash
cd ai/kvcache/mock
./run.sh
```

2. Send manual requests:

```bash
./request.sh
```

3. Run test case:

```bash
go test -v ./test/pixiu_test.go
```

## Verification Targets

- tokenize call works (`engine-a`)
- lookup/pin calls work (`controller`)
- second same prompt is routed to preferred endpoint (`mock-llm-b`)
