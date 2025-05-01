// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

import (
	"github.com/redis/go-redis/v9"
)

var (
	// ErrNil 当 Get、GetStruct 等方法期望返回非 nil 值但实际返回 nil 时返回此错误。
	ErrNil = redis.Nil

	// ErrClosed 当对已关闭的客户端执行操作时返回此错误。
	ErrClosed = redis.ErrClosed

	// TxFailedErr 表示 Redis 事务执行失败的错误。
	TxFailedErr = redis.TxFailedErr //nolint:errname
)

// 基础命令接口定义。
type (
	// Cmder 表示 Redis 命令的基础接口。
	Cmder = redis.Cmder

	// Pipeliner 表示 Redis 管道操作接口。
	Pipeliner = redis.Pipeliner

	// Scripter 表示 Redis 脚本操作接口。
	Scripter = redis.Scripter
)

// 命令类型定义。
type (
	// Cmd 表示通用的 Redis 命令。
	Cmd = redis.Cmd

	// SliceCmd 表示返回切片类型的 Redis 命令。
	SliceCmd = redis.SliceCmd

	// StatusCmd 表示返回状态类型的 Redis 命令。
	StatusCmd = redis.StatusCmd

	// IntCmd 表示返回整数类型的 Redis 命令。
	IntCmd = redis.IntCmd

	// IntSliceCmd 表示返回整数切片类型的 Redis 命令。
	IntSliceCmd = redis.IntSliceCmd

	// DurationCmd 表示返回时间间隔类型的 Redis 命令。
	DurationCmd = redis.DurationCmd

	// TimeCmd 表示返回时间类型的 Redis 命令。
	TimeCmd = redis.TimeCmd

	// BoolCmd 表示返回布尔类型的 Redis 命令。
	BoolCmd = redis.BoolCmd

	// StringCmd 表示返回字符串类型的 Redis 命令。
	StringCmd = redis.StringCmd

	// FloatCmd 表示返回浮点数类型的 Redis 命令。
	FloatCmd = redis.FloatCmd

	// FloatSliceCmd 表示返回浮点数切片类型的 Redis 命令。
	FloatSliceCmd = redis.FloatSliceCmd

	// StringSliceCmd 表示返回字符串切片类型的 Redis 命令。
	StringSliceCmd = redis.StringSliceCmd

	// KeyValue 表示键值对类型。
	KeyValue = redis.KeyValue

	// KeyValueSliceCmd 表示返回键值对切片类型的 Redis 命令。
	KeyValueSliceCmd = redis.KeyValueSliceCmd

	// BoolSliceCmd 表示返回布尔切片类型的 Redis 命令。
	BoolSliceCmd = redis.BoolSliceCmd

	// MapStringStringCmd 表示返回字符串到字符串映射类型的 Redis 命令。
	MapStringStringCmd = redis.MapStringStringSliceCmd

	// MapStringIntCmd 表示返回字符串到整数映射类型的 Redis 命令。
	MapStringIntCmd = redis.MapStringIntCmd

	// StringStructMapCmd 表示返回字符串到结构体映射类型的 Redis 命令。
	StringStructMapCmd = redis.StringStructMapCmd

	// XMessage 表示 Redis Stream 消息类型。
	XMessage = redis.XMessage

	// XMessageSliceCmd 表示返回 Stream 消息切片类型的 Redis 命令。
	XMessageSliceCmd = redis.XMessageSliceCmd

	// XStream 表示 Redis Stream 类型。
	XStream = redis.XStream

	// XStreamSliceCmd 表示返回 Stream 切片类型的 Redis 命令。
	XStreamSliceCmd = redis.XStreamSliceCmd

	// XPending 表示待处理消息类型。
	XPending = redis.XPending

	// XPendingCmd 表示返回待处理消息类型的 Redis 命令。
	XPendingCmd = redis.XPendingCmd

	// XPendingExt 表示扩展的待处理消息类型。
	XPendingExt = redis.XPendingExt

	// XPendingExtCmd 表示返回扩展待处理消息类型的 Redis 命令。
	XPendingExtCmd = redis.XPendingExtCmd

	// XAutoClaimCmd 表示自动认领消息的 Redis 命令。
	XAutoClaimCmd = redis.XAutoClaimCmd

	// XAutoClaimJustIDCmd 表示仅返回自动认领消息 ID 的 Redis 命令。
	XAutoClaimJustIDCmd = redis.XAutoClaimJustIDCmd

	// XInfoConsumersCmd 表示返回消费者信息的 Redis 命令。
	XInfoConsumersCmd = redis.XInfoConsumersCmd

	// XInfoConsumer 表示消费者信息类型。
	XInfoConsumer = redis.XInfoConsumer

	// XInfoGroupsCmd 表示返回消费者组信息的 Redis 命令。
	XInfoGroupsCmd = redis.XInfoGroupsCmd

	// XInfoGroup 表示消费者组信息类型。
	XInfoGroup = redis.XInfoGroup

	// XInfoStreamCmd 表示返回 Stream 信息的 Redis 命令。
	XInfoStreamCmd = redis.XInfoStreamCmd

	// XInfoStream 表示 Stream 信息类型。
	XInfoStream = redis.XInfoStream

	// XInfoStreamFullCmd 表示返回完整 Stream 信息的 Redis 命令。
	XInfoStreamFullCmd = redis.XInfoStreamFullCmd

	// XInfoStreamFull 表示完整 Stream 信息类型。
	XInfoStreamFull = redis.XInfoStreamFull

	// XInfoStreamGroup 表示 Stream 组信息类型。
	XInfoStreamGroup = redis.XInfoStreamGroup

	// XInfoStreamGroupPending 表示 Stream 组待处理消息信息类型。
	XInfoStreamGroupPending = redis.XInfoStreamGroupPending

	// XInfoStreamConsumer 表示 Stream 消费者信息类型。
	XInfoStreamConsumer = redis.XInfoStreamConsumer

	// XInfoStreamConsumerPending 表示 Stream 消费者待处理消息信息类型。
	XInfoStreamConsumerPending = redis.XInfoStreamConsumerPending

	// ZSliceCmd 表示返回有序集合切片类型的 Redis 命令。
	ZSliceCmd = redis.ZSliceCmd

	// ZWithKeyCmd 表示返回带键的有序集合类型的 Redis 命令。
	ZWithKeyCmd = redis.ZWithKeyCmd

	// ScanCmd 表示扫描操作的 Redis 命令。
	ScanCmd = redis.ScanCmd

	// ClusterNode 表示集群节点信息类型。
	ClusterNode = redis.ClusterNode

	// ClusterSlot 表示集群槽位信息类型。
	ClusterSlot = redis.ClusterSlot

	// ClusterSlotsCmd 表示返回集群槽位信息的 Redis 命令。
	ClusterSlotsCmd = redis.ClusterSlotsCmd

	// GeoLocation 表示地理位置信息类型。
	GeoLocation = redis.GeoLocation

	// GeoRadiusQuery 表示地理位置半径查询类型。
	GeoRadiusQuery = redis.GeoRadiusQuery

	// GeoLocationCmd 表示返回地理位置信息的 Redis 命令。
	GeoLocationCmd = redis.GeoLocationCmd

	// GeoSearchQuery 表示地理位置搜索查询类型。
	GeoSearchQuery = redis.GeoSearchQuery

	// GeoSearchLocationQuery 表示地理位置搜索位置查询类型。
	GeoSearchLocationQuery = redis.GeoSearchLocationQuery

	// GeoSearchStoreQuery 表示地理位置搜索存储查询类型。
	GeoSearchStoreQuery = redis.GeoSearchStoreQuery

	// GeoSearchLocationCmd 表示返回地理位置搜索结果的 Redis 命令。
	GeoSearchLocationCmd = redis.GeoSearchLocationCmd

	// GeoPos 表示地理位置坐标类型。
	GeoPos = redis.GeoPos

	// GeoPosCmd 表示返回地理位置坐标的 Redis 命令。
	GeoPosCmd = redis.GeoPosCmd

	// CommandInfo 表示命令信息类型。
	CommandInfo = redis.CommandInfo

	// CommandsInfoCmd 表示返回命令信息的 Redis 命令。
	CommandsInfoCmd = redis.CommandsInfoCmd

	// SlowLog 表示慢查询日志类型。
	SlowLog = redis.SlowLog

	// SlowLogCmd 表示返回慢查询日志的 Redis 命令。
	SlowLogCmd = redis.SlowLogCmd

	// MapStringInterfaceCmd 表示返回字符串到接口映射类型的 Redis 命令。
	MapStringInterfaceCmd = redis.MapStringInterfaceCmd

	// MapStringStringSliceCmd 表示返回字符串到字符串切片映射类型的 Redis 命令。
	MapStringStringSliceCmd = redis.MapStringStringSliceCmd

	// KeyValuesCmd 表示返回键值对类型的 Redis 命令。
	KeyValuesCmd = redis.KeyValuesCmd

	// ZSliceWithKeyCmd 表示返回带键的有序集合切片类型的 Redis 命令。
	ZSliceWithKeyCmd = redis.ZSliceWithKeyCmd

	// Function 表示 Redis 函数类型。
	Function = redis.Function

	// Library 表示 Redis 库类型。
	Library = redis.Library

	// FunctionListCmd 表示返回函数列表的 Redis 命令。
	FunctionListCmd = redis.FunctionListCmd

	// FunctionStats 表示函数统计信息类型。
	FunctionStats = redis.FunctionStats

	// RunningScript 表示运行中的脚本类型。
	RunningScript = redis.RunningScript

	// Engine 表示脚本引擎类型。
	Engine = redis.Engine

	// FunctionStatsCmd 表示返回函数统计信息的 Redis 命令。
	FunctionStatsCmd = redis.FunctionStatsCmd

	// LCSQuery 表示最长公共子序列查询类型。
	LCSQuery = redis.LCSQuery

	// LCSMatch 表示最长公共子序列匹配类型。
	LCSMatch = redis.LCSMatch

	// LCSMatchedPosition 表示最长公共子序列匹配位置类型。
	LCSMatchedPosition = redis.LCSMatchedPosition

	// LCSPosition 表示最长公共子序列位置类型。
	LCSPosition = redis.LCSPosition

	// LCSCmd 表示返回最长公共子序列的 Redis 命令。
	LCSCmd = redis.LCSCmd

	// KeyFlags 表示键标志类型。
	KeyFlags = redis.KeyFlags

	// KeyFlagsCmd 表示返回键标志的 Redis 命令。
	KeyFlagsCmd = redis.KeyFlagsCmd

	// ClusterLink 表示集群链接类型。
	ClusterLink = redis.ClusterLink

	// ClusterLinksCmd 表示返回集群链接信息的 Redis 命令。
	ClusterLinksCmd = redis.ClusterLinksCmd

	// SlotRange 表示槽位范围类型。
	SlotRange = redis.SlotRange

	// Node 表示节点类型。
	Node = redis.Node

	// ClusterShard 表示集群分片类型。
	ClusterShard = redis.ClusterShard

	// ClusterShardsCmd 表示返回集群分片信息的 Redis 命令。
	ClusterShardsCmd = redis.ClusterShardsCmd

	// RankScore 表示排名分数类型。
	RankScore = redis.RankScore

	// RankWithScoreCmd 表示返回带分数的排名的 Redis 命令。
	RankWithScoreCmd = redis.RankWithScoreCmd

	// ClientFlags 表示客户端标志类型。
	ClientFlags = redis.ClientFlags

	// ClientInfo 表示客户端信息类型。
	ClientInfo = redis.ClientInfo

	// ClientInfoCmd 表示返回客户端信息的 Redis 命令。
	ClientInfoCmd = redis.ClientInfoCmd

	// ACLLogEntry 表示 ACL 日志条目类型。
	ACLLogEntry = redis.ACLLogEntry

	// ACLLogCmd 表示返回 ACL 日志的 Redis 命令。
	ACLLogCmd = redis.ACLLogCmd
)

// 命令参数类型定义。
type (
	// FilterBy 表示过滤条件类型。
	FilterBy = redis.FilterBy

	// Sort 表示排序参数类型。
	Sort = redis.Sort

	// SetArgs 表示设置命令的参数类型。
	SetArgs = redis.SetArgs

	// BitCount 表示位计数参数类型。
	BitCount = redis.BitCount

	// LPosArgs 表示列表位置查询参数类型。
	LPosArgs = redis.LPosArgs

	// XAddArgs 表示 Stream 添加消息的参数类型。
	XAddArgs = redis.XAddArgs

	// XReadArgs 表示 Stream 读取消息的参数类型。
	XReadArgs = redis.XReadArgs

	// XReadGroupArgs 表示 Stream 消费者组读取消息的参数类型。
	XReadGroupArgs = redis.XReadGroupArgs

	// XPendingExtArgs 表示扩展待处理消息查询参数类型。
	XPendingExtArgs = redis.XPendingExtArgs

	// XClaimArgs 表示消息认领参数类型。
	XClaimArgs = redis.XClaimArgs

	// Z 表示有序集合元素类型。
	Z = redis.Z

	// ZWithKey 表示带键的有序集合元素类型。
	ZWithKey = redis.ZWithKey

	// ZStore 表示有序集合存储参数类型。
	ZStore = redis.ZStore

	// ZAddArgs 表示有序集合添加元素的参数类型。
	ZAddArgs = redis.ZAddArgs

	// ZRangeArgs 表示有序集合范围查询参数类型。
	ZRangeArgs = redis.ZRangeArgs

	// ZRangeBy 表示有序集合范围查询参数类型。
	ZRangeBy = redis.ZRangeBy

	// FunctionListQuery 表示函数列表查询参数类型。
	FunctionListQuery = redis.FunctionListQuery

	// ModuleLoadexConfig 表示模块加载配置类型。
	ModuleLoadexConfig = redis.ModuleLoadexConfig
)

// 发布订阅相关类型定义。
type (
	// PubSub 表示发布订阅客户端类型。
	PubSub = redis.PubSub

	// Subscription 表示订阅信息类型。
	Subscription = redis.Subscription

	// Message 表示消息类型。
	Message = redis.Message

	// Pong 表示 PING 命令响应类型。
	Pong = redis.Pong
)
