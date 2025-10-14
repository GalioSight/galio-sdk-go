// Copyright 2024 Tencent Galileo Authors

// Package attribute ...
package attribute

// attribute 去重问题的说明
// 1. tracer.Start(..., trace.WithAttributes(...))
//   这部分 attribute 在 sampler 使用时不会去重
// 2. span.SetAttribute(...)
//   这部分 attribute 在设置时不会去重
// 3. span.Attributes()
//   在观测 attribute 时会去重，并更新回 span.attribute
// 因为 1 内部基本不会有重复 attr 的需求，所以没有关系
// 1 和 2 在观测时会自动去重
