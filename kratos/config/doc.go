// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package config 提供了一个灵活的配置解码器，用于处理 Kratos 框架中的配置管理。

主要功能：

1. 配置解码：
  - 支持点分隔键的配置展开
  - 支持自定义解析函数
  - 支持配置验证
  - 支持默认值设置

2. 解码器选项：
  - WithResolve：设置自定义解析函数
  - WithValidator：设置配置验证器
  - WithDefaults：设置默认值

基本用法：

1. 创建解码器：

	// 创建基本解码器
	decoder := NewDecoder()

	// 创建带选项的解码器
	decoder := NewDecoder(
	    WithResolve(func(target map[string]interface{}) error {
	        // 自定义解析逻辑
	        return nil
	    }),
	)

2. 解码配置：

	var cfg struct {
	    Database struct {
	        Host string `json:"host"`
	        Port int    `json:"port"`
	    } `json:"database"`
	}

	// 解码配置到结构体
	err := decoder.Decode(kv, &cfg)

3. 自定义解析：

	decoder := NewDecoder(
	    WithResolve(func(target map[string]interface{}) error {
	        // 处理特殊配置键
	        if v, ok := target["special.key"]; ok {
	            // 自定义处理逻辑
	        }
	        return nil
	    }),
	)

配置格式：

1. 点分隔键：
  - 支持将 "a.b.c" 形式的键展开为嵌套映射
  - 自动处理数组索引
  - 支持复杂数据结构

2. 值类型：
  - 支持基本数据类型
  - 支持复合数据类型
  - 支持自定义类型转换

性能优化：

1. 缓存机制：
  - 缓存解析结果
  - 减少重复解析
  - 优化内存使用

2. 并发处理：
  - 支持并发安全的配置读取
  - 避免锁竞争
  - 优化性能

使用建议：

1. 配置结构：
  - 使用清晰的配置层次
  - 避免过深的嵌套
  - 合理组织配置项

2. 错误处理：
  - 检查解码错误
  - 验证配置有效性
  - 提供合理的默认值

3. 性能考虑：
  - 避免频繁解码
  - 合理使用缓存
  - 注意内存使用

注意事项：

1. 类型转换：
  - 注意类型兼容性
  - 处理类型转换错误
  - 提供类型转换函数

2. 配置验证：
  - 验证必需字段
  - 检查值范围
  - 验证数据格式

3. 并发安全：
  - 注意配置读写并发
  - 使用适当的锁机制
  - 避免竞态条件

更多示例请参考 example/config 目录。
*/
package config
