// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
//
// 2024 Tencent Galileo Authors

package main

import (
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer trace.Tracer
	meter  metric.Meter
	logger *slog.Logger

	rollCnt metric.Int64Counter
)

func doInit() {
	tracer = otel.Tracer("rolldice")
	meter = otel.Meter("rolldice")
	logger = otelslog.NewLogger("rolldice")
	var err error
	rollCnt, err = meter.Int64Counter("dice.rolls",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{roll}"))
	if err != nil {
		panic(err)
	}
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "roll")
	defer span.End()

	roll := 1 + rand.Intn(6)

	logger.InfoContext(ctx, "Anonymous player is rolling the dice", "result", roll)

	rollValueAttr := attribute.Int("roll.value", roll)
	span.SetAttributes(rollValueAttr)
	rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	resp := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		logger.ErrorContext(ctx, "Write failed", "error", err)
	}
}
