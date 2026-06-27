// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package validate 提供用于 Kratos 服务端的请求校验中间件。
//
// Validator 只对实现 Validate() error 的请求对象执行校验；未实现该方法的
// 请求会直接透传给后续处理器。校验失败时，默认回调会返回 code 为
// VALIDATOR 的 BadRequest，并把原始校验错误保存在 cause 中。
//
// 调用方可以通过 WithValidateCallback 自定义错误转换逻辑，或为校验失败
// 返回替代响应。
package validate
