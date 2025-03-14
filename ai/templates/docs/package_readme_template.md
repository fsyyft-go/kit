# {{.PackageName}}

## 简介

{{.PackageDescription}}

### 主要特性

{{range .Features}}
- {{.}}
{{end}}

### 设计理念

{{.DesignPhilosophy}}

## 安装

### 前置条件

- Go 版本要求：{{.MinGoVersion}}
- 依赖要求：
{{range .Dependencies}}
  - {{.Name}} v{{.Version}}
{{end}}

### 安装命令

```bash
go get -u {{.PackageImportPath}}
```

## 快速开始

### 基础用法

```go
{{.BasicExample}}
```

### 配置选项

```go
{{.ConfigExample}}
```

## 详细指南

### 核心概念

{{.CoreConcepts}}

### 常见用例

#### 1. {{.UseCase1.Title}}

```go
{{.UseCase1.Code}}
```

#### 2. {{.UseCase2.Title}}

```go
{{.UseCase2.Code}}
```

### 最佳实践

{{range .BestPractices}}
- {{.}}
{{end}}

## API 文档

### 主要类型

```go
{{.MainTypes}}
```

### 关键函数

{{range .KeyFunctions}}
#### {{.Name}}

{{.Description}}

```go
{{.Signature}}
```

示例：
```go
{{.Example}}
```
{{end}}

### 错误处理

{{.ErrorHandling}}

## 性能指标

{{if .PerformanceMetrics}}
| 操作 | 性能指标 | 说明 |
|------|----------|------|
{{range .PerformanceMetrics}}
| {{.Operation}} | {{.Metric}} | {{.Description}} |
{{end}}
{{else}}
性能测试结果待补充
{{end}}

## 测试覆盖率

{{if .TestCoverage}}
| 包 | 覆盖率 |
|------|--------|
{{range .TestCoverage}}
| {{.Package}} | {{.Coverage}} |
{{end}}
{{else}}
测试覆盖率信息待补充
{{end}}

## 调试指南

### 日志级别

{{range .LogLevels}}
- {{.Level}}: {{.Description}}
{{end}}

### 常见问题排查

{{range .Troubleshooting}}
#### {{.Problem}}

{{.Solution}}
{{end}}

## 相关文档

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南]({{.ContributingGuideLink}})了解详细信息。

## 许可证

本项目采用 {{.License}} 许可证。查看 [LICENSE](./LICENSE) 文件了解更多信息。

## 补充说明

本文的大部分信息，由 AI 使用[模板](package_readme_template.md)根据[提示词](../../prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。