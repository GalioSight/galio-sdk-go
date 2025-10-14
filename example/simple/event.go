// Copyright 2025 Tencent Galileo Authors
//
// Copyright 2025 Tencent OpenTelemetry Oteam
//
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main ...
package main

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"galiosight.ai/galio-sdk-go"
	logconf "galiosight.ai/galio-sdk-go/configs/logs"
	"galiosight.ai/galio-sdk-go/model"
)

func eventsDemo(resource *model.Resource) {
	// 初始化，只能执行一次
	logsConfig := logconf.NewConfig(resource)
	logsConfig.Processor.Level = "INFO"
	logger, err := galio.NewLogger(logsConfig) // 全局持有，不要重复创建。
	if err != nil {
		fmt.Printf("NewLogger err=%v, logsConfig=%+v", err, logsConfig)
		return
	}
	galio.SetEventLogger(logger)
	// 以下是事件上报示例
	for {
		// 事件示例 1: 简单事件
		galio.ReportEvent("事件详细描述：example detail", "galileo", "exampleDomain", "业务配置变更")
		// 事件示例 2：缓存定时加载事件
		galio.ReportEvent(
			"缓存定时加载成功，加载的缓存项数量：100",
			"缓存管理系统",
			"后台服务",
			"缓存加载",
			zap.Int("加载项数量", 100),                              // 加载的缓存项数量
			zap.String("缓存类型", "物品信息"),                         // 缓存的类型
			zap.String("加载状态", "成功"),                           // 加载状态
			zap.Float64("加载耗时", 1.23),                          // 加载耗时（单位：秒）
			zap.Time("加载开始时间", time.Now().Add(-time.Second*2)), // 加载开始时间
			zap.Time("加载结束时间", time.Now()),                     // 加载结束时间
			zap.String("下次加载时间", "2024-12-18 10:00:00"),        // 下次加载的计划时间
		)
		go func() {
			defer Recover()
			panic("panic")
		}()
		time.Sleep(time.Minute)
	}
}
