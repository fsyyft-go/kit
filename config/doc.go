// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package config 提供基于构建上下文的版本信息访问与格式化输出。
//
// 本包当前围绕 CurrentVersion 暴露应用和类库的版本号、Git 提交、构建时间以及构建目
// 录信息，并透传底层 BuildingContext 的调试状态。它不负责通用配置加载；包名中的
// config 目前主要承载版本元数据展示能力。
//
// CurrentVersion 是默认构建信息实例。
//   - 它作为值可直接参与 fmt 格式化输出；默认格式返回简短版本串 version
//     <git-short>/<build-time> (build <go-version>)，使用 %+v 时返回 Description
//     生成的多行诊断文本。
//   - 由于 CurrentVersion 是包级变量，调用方可直接调用 CurrentVersion.Version()、
//     CurrentVersion.GitVersion()、CurrentVersion.BuildTimeString() 和
//     CurrentVersion.Description() 获取版本与构建信息。
//   - 当需要按 fmt.Stringer 或 github.com/fsyyft-go/kit/go/build.BuildingContext
//     接口传递版本信息时，应使用 &CurrentVersion；相关访问器以及 String 和
//     Description 方法由 *version 实现，并直接透传底层构建字段。
//
// Description 会输出 Go 版本、编译时间、应用与类库的 Git 版本，以及类库目录、工作目
// 录、GOROOT 和 GOPATH，适合启动日志、诊断页面和构建排障场景。
package config
