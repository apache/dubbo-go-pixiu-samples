# KVCache Mock 示例

[返回 KVCache 索引](../README_zh.md) | [English](./README.md) | 中文

## 目录结构

- `pixiu/conf.yaml`: Pixiu 配置
- `server/controller/main.go`: mock LMCache controller（`/lookup` `/pin` `/compress` `/evict`）
- `server/engine-a/main.go`: mock 引擎 A（`/tokenize` + `/v1/chat/completions`）
- `server/engine-b/main.go`: mock 引擎 B（`/v1/chat/completions`）
- `request.sh`: 请求脚本
- `test/pixiu_test.go`: 集成测试
- `run.sh`: 一键启动并验收

## 命令行快速开始

1. 启动 sample 服务和 Pixiu：

```bash
cd ai/kvcache/mock
./run.sh
```

2. 手工发请求：

```bash
./request.sh
```

3. 运行测试用例：

```bash
go test -v ./test/pixiu_test.go
```

## 验证目标

- tokenize 调用生效（`engine-a`）
- lookup/pin 调用生效（`controller`）
- 同 prompt 第二次请求路由到 preferred endpoint（`mock-llm-b`）
