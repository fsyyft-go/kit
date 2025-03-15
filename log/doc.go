// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package log 提供了一个统一的日志接口和多种日志实现。

主要特性：

  - 支持多种日志后端（标准输出、Logrus）
  - 提供统一的日志接口
  - 支持结构化日志记录
  - 支持多个日志级别
  - 支持文件和标准输出
  - 支持函数式配置选项
  - 支持日志文件轮转
  - 支持日志格式化（文本/JSON）

日志级别：

  - DebugLevel：调试信息，仅在开发环境使用
  - InfoLevel：正常操作信息，用于跟踪应用状态
  - WarnLevel：警告信息，表示潜在问题
  - ErrorLevel：错误信息，表示操作失败
  - FatalLevel：致命错误，记录后程序退出

基本使用：

	// 使用默认配置初始化日志
	if err := log.InitLogger(); err != nil {
	    panic(err)
	}

	// 使用自定义配置初始化日志
	if err := log.InitLogger(
	    log.WithLogType(log.LogTypeLogrus),
	    log.WithLevel(log.DebugLevel),
	    log.WithOutput("/var/log/app.log"),
	    log.WithFormatType(log.JSONFormat),
	); err != nil {
	    panic(err)
	}

	// 记录不同级别的日志
	log.Debug("调试信息")
	log.Info("正常信息")
	log.Warn("警告信息")
	log.Error("错误信息")

	// 使用结构化日志
	log.WithField("user", "admin").Info("用户登录")
	log.WithFields(map[string]interface{}{
	    "user":   "admin",
	    "action": "login",
	    "time":   time.Now(),
	}).Info("用户操作")

日志轮转：

	// 启用日志轮转
	if err := log.InitLogger(
	    log.WithOutput("/var/log/app.log"),
	    log.WithEnableRotate(true),
	    log.WithRotateTime(24 * time.Hour),
	    log.WithMaxAge(7 * 24 * time.Hour),
	); err != nil {
	    panic(err)
	}

独立日志实例：

	// 创建独立的日志实例
	logger, err := log.NewLogger(
	    log.WithLogType(log.LogTypeStd),
	    log.WithLevel(log.DebugLevel),
	)
	if err != nil {
	    panic(err)
	}
	defer logger.Close()

	// 使用独立实例记录日志
	logger.Info("使用独立的日志实例")

更多示例请参考 example/log 目录。
*/
package log
