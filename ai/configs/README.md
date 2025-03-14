# AI 配置文件

本目录包含 AI 工具的配置文件。

## 配置文件说明

- `openai.yaml`: OpenAI API 的配置文件
- `anthropic.yaml`: Anthropic API 的配置文件

## 配置文件格式

配置文件应使用 YAML 格式，包含以下内容：

```yaml
# API 配置示例
api:
  # API 密钥（通过环境变量获取）
  key: ${API_KEY}
  # API 基础 URL
  base_url: https://api.example.com
  # 默认模型设置
  model: gpt-4
  # 请求超时设置（秒）
  timeout: 30

# 模型参数
parameters:
  temperature: 0.7
  max_tokens: 2000
  top_p: 1.0

# 自定义设置
custom:
  retry_count: 3
  cache_enabled: true
```

## 使用指南

1. 复制示例配置文件为 `config.yaml`
2. 设置必要的环境变量
3. 根据需要调整配置参数

## 安全注意事项

- 不要在配置文件中直接存储 API 密钥
- 使用环境变量存储敏感信息
- 确保配置文件不被提交到版本控制系统
- 定期更新和审查配置 