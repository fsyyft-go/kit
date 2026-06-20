// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package driver 提供 database/sql/driver 的包装驱动和 Hook 基础设施。
//
// 本包通过 NewKitDriver 将底层 driver.Driver 包装为可插入 Hook 的实现，
// 使连接、预处理、执行、查询、Ping、事务开始、提交和回滚等阶段都可以
// 在调用前后同步执行自定义逻辑。
//
// HookContext 记录操作类型、SQL、参数、耗时、原始结果和原始错误，并实现
// context.Context 以便 Hook 共享取消信号和上下文值。HookManager 按注册顺序
// 执行 Before、按逆序执行 After；NewHookLogError 和 NewHookLogSlow 则提供
// 错误日志与慢查询日志的现成 Hook。
//
// 本包只负责驱动包装与 Hook 编排，不负责注册具体数据库驱动或创建 *sql.DB。
package driver
