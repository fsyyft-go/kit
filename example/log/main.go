package main

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/fsyyft-go/kit/log"
)

func main() {
	// 示例1：使用标准输出日志器。
	if err := log.InitLogger(log.LogTypeConsole, ""); err != nil {
		panic(err)
	}

	// 设置日志级别。
	log.SetLevel(log.DebugLevel)

	// 基本日志记录。
	log.Debug("这是一条调试日志")
	log.Info("这是一条信息日志")
	log.Warn("这是一条警告日志")
	log.Error("这是一条错误日志")

	// 格式化日志。
	log.Debugf("当前时间是: %v", time.Now().Format("2006-01-02 15:04:05"))
	log.Infof("程序运行在: %s", os.Getenv("PWD"))

	// 结构化日志。
	log.WithField("user", "admin").Info("用户登录")
	log.WithFields(map[string]interface{}{
		"ip":      "192.168.1.1",
		"method":  "POST",
		"latency": "20ms",
	}).Info("收到HTTP请求")

	// 错误处理示例。
	if err := someFunction(); err != nil {
		log.WithField("error", err).Error("操作失败")
	}

	// 示例2：使用 logrus 日志器。
	logFile := filepath.Join("example", "log", "app.log")
	if err := log.InitLogger(log.LogTypeLogrus, logFile); err != nil {
		panic(err)
	}

	// 使用新的日志器记录。
	log.Info("已切换到 logrus 日志器")
	log.WithFields(map[string]interface{}{
		"component": "server",
		"status":    "starting",
	}).Info("服务器启动")
}

func someFunction() error {
	return errors.New("示例错误")
}
