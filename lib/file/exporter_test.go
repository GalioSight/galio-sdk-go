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

package file

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
	pprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
)

func Test_fileExporter_Export(t *testing.T) {
	type fields struct {
		exportToFile bool
		path         string
		log          *logs.Wrapper
	}
	type expected struct {
		fileCnt     int
		fileContent string
	}
	type args struct {
		m interface{}
	}
	const path = "file_exporter"
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected expected
	}{
		{
			name: "导出",
			fields: fields{
				exportToFile: true,
				path:         path,
				log:          logs.DefaultWrapper(),
			},
			args: args{
				m: map[string]string{
					"a": "b",
				},
			},
			expected: expected{
				fileCnt:     1,
				fileContent: `{"a":"b"}`,
			},
		},
		{
			name: "不导出",
			fields: fields{
				exportToFile: false,
				path:         path,
				log:          logs.DefaultWrapper(),
			},
			args: args{
				m: map[string]string{
					"a": "b",
				},
			},
			expected: expected{},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				f := NewExporter(tt.fields.exportToFile, tt.fields.path, tt.fields.log)
				f.Export(tt.args.m)
				if tt.fields.exportToFile {
					assert.DirExists(t, path)
					readDir, err := ioutil.ReadDir(path)
					assert.Nil(t, err)
					assert.NotNil(t, readDir)
					assert.Equal(t, tt.expected.fileCnt, len(readDir))
					for i := 0; i < tt.expected.fileCnt; i++ {
						bytes, err := ioutil.ReadFile(path + "/" + readDir[i].Name())
						assert.Nil(t, err)
						assert.Equal(t, tt.expected.fileContent, string(bytes))
					}
					err = os.RemoveAll(path)
					assert.Nil(t, err)
				} else {
					assert.NoDirExists(t, path)
				}
			},
		)
	}
}

func Test_fileExporter_ExportProfilesBatch(t *testing.T) {
	prof := &pprofile.Profile{
		SampleType: []*pprofile.ValueType{
			{
				Type: "contention",
				Unit: "count",
			},
			{
				Type: "delay",
				Unit: "nanoseconeds",
			},
		},
		Sample: []*pprofile.Sample{
			{
				Location: []*pprofile.Location{
					{
						ID: 1,
						Mapping: &pprofile.Mapping{
							ID:           1,
							HasFunctions: true,
						},
						Address: 1,
						Line: []pprofile.Line{
							{
								Function: &pprofile.Function{
									ID:   1,
									Name: "main",
								},
								Line: 0,
							},
						},
					},
				},
				Value: []int64{3, 1},
			},
		},
		Mapping: []*pprofile.Mapping{
			{
				ID:           1,
				HasFunctions: true,
			},
		},
		Location: []*pprofile.Location{
			{
				ID: 1,
				Mapping: &pprofile.Mapping{
					ID:           1,
					HasFunctions: true,
				},
				Address: 1,
				Line: []pprofile.Line{
					{
						Function: &pprofile.Function{
							ID:   1,
							Name: "main",
						},
						Line: 0,
					},
				},
			},
		},
		Function: []*pprofile.Function{
			{
				ID:   1,
				Name: "main",
			},
		},
	}
	var buf bytes.Buffer
	if err := prof.Write(&buf); err != nil {
		t.Fatal(err)
	}
	batch := &model.ProfilesBatch{
		Sequence: 1,
		Start:    1672502400,
		End:      1672502460,
		Profiles: []*model.Profile{
			{
				Name: "mutex.pprof",
				Type: "mutex",
				Data: buf.Bytes(),
			},
		},
	}
	type fields struct {
		exportToFile bool
		path         string
		log          *logs.Wrapper
	}
	type expected struct {
		fileCnt     int
		fileContent *pprofile.Profile
	}
	const path = "file_exporter"
	tests := []struct {
		name     string
		fields   fields
		args     *model.ProfilesBatch
		expected expected
	}{
		{
			name: "导出",
			fields: fields{
				exportToFile: true,
				path:         path,
				log:          logs.DefaultWrapper(),
			},
			args: batch,
			expected: expected{
				fileCnt:     1,
				fileContent: prof.Copy(),
			},
		},
		{
			name: "不导出",
			fields: fields{
				exportToFile: false,
				path:         path,
				log:          logs.DefaultWrapper(),
			},
			args:     batch,
			expected: expected{},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				f := NewExporter(tt.fields.exportToFile, tt.fields.path, tt.fields.log)
				f.ExportProfilesBatch(tt.args)
				if tt.fields.exportToFile {
					dir := filepath.Join(path, strconv.FormatInt(tt.args.End, 10))
					assert.DirExists(t, dir)
					readDir, err := ioutil.ReadDir(dir)
					assert.Nil(t, err)
					assert.NotNil(t, readDir)
					assert.Equal(t, tt.expected.fileCnt, len(readDir))
					for i := 0; i < tt.expected.fileCnt; i++ {
						bytes, err := ioutil.ReadFile(dir + "/" + readDir[i].Name())
						assert.Nil(t, err)
						got, err := pprofile.ParseData(bytes)
						if err != nil {
							t.Error("cannot parse bytes data to profile format")
						}
						assert.True(t, reflect.DeepEqual(got, tt.expected.fileContent))
					}
					err = os.RemoveAll(path)
					assert.Nil(t, err)
				} else {
					assert.NoDirExists(t, path)
				}
			},
		)
	}
}
