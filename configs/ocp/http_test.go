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

package ocp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	s, err := post(
		DefaultURL, []byte("{}"), timeout, map[string]string{
			"Content-Type": "application/json",
		},
	)
	assert.Nil(t, err)
	assert.True(t, len(s) > 0)
}

func TestQueryError(t *testing.T) {
	s, err := post(
		"s", []byte(""), timeout, map[string]string{
			"Content-Type": "application/json",
		},
	)
	assert.NotNil(t, err)
	assert.True(t, len(s) == 0)
}
