# KVCache BYOE（真实引擎）示例

[返回 KVCache 索引](../README_zh.md) | [English](./README.md) | 中文

## 目录结构

- `pixiu/conf.yaml`: BYOE Pixiu 模板（环境变量驱动）
- `request.sh`: 手工请求脚本
- `test/pixiu_test.go`: 冒烟测试（依赖外部 BYOE 环境）
- `verify.sh`: 指标验证脚本

## 前置条件

- 可访问的 `VLLM_ENDPOINT`（包含 `/tokenize`）
- 可访问的 `LMCACHE_ENDPOINT`（包含 `/lookup` `/pin` `/compress` `/evict`）
- 可用 Pixiu 二进制或 Pixiu 源码仓库

## 命令行快速开始

1. 配置环境变量并执行验证：

```bash
cd ai/kvcache/real-engine
export VLLM_ENDPOINT="http://<vllm-host>:<port>"
export LMCACHE_ENDPOINT="http://<lmcache-host>:<port>"
./verify.sh
```

2. 手工发请求：

```bash
./request.sh
```

3. 运行测试用例：

```bash
go test -v ./test/pixiu_test.go
```

## 验收指标

- lookup 成功率
- preferred endpoint 命中率
- p95 延迟对比（无缓存 vs 有缓存）
