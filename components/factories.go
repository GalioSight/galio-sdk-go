// Copyright 2021 Tencent Galileo Authors
//
// Copyright 2021 Tencent OpenTelemetry Oteam
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

package components

// Factories 工厂总表（处理器 + 导出器）。
// 具体选择哪个工厂，由用户配置驱动。
type Factories struct {
	ProcessorFactories map[string]ProcessorFactory // 协议 -> 处理器工厂。
	ExporterFactories  map[string]ExporterFactory  // 协议 -> 导出器工厂。
}

// BuildProcessorFactories 输入处理器工厂列表，输出协议到处理器工厂的映射。
func BuildProcessorFactories(factories ...ProcessorFactory) map[string]ProcessorFactory {
	processorFactories := make(map[string]ProcessorFactory, len(factories))
	for _, f := range factories {
		if _, ok := processorFactories[f.Protocol()]; ok {
			continue
		}
		processorFactories[f.Protocol()] = f
	}
	return processorFactories
}

// BuildExporterFactories 输入导出器工厂列表，输出协议到导出器工厂的映射。
func BuildExporterFactories(factories ...ExporterFactory) map[string]ExporterFactory {
	exporterFactories := make(map[string]ExporterFactory, len(factories))
	for _, f := range factories {
		if _, ok := exporterFactories[f.Protocol()]; ok {
			continue
		}
		exporterFactories[f.Protocol()] = f
	}
	return exporterFactories
}
