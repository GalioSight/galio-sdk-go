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
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"galiosight.ai/galio-sdk-go/lib/logs"
	"galiosight.ai/galio-sdk-go/model"
)

// Exporter 文件导出器，将指标数据导出到文件。
// 每次导出都输出到一个文件中，文件名 $path/$(UnixNano).json.
// 一般情况下，不会出现 UnixNano 相同，从而互相覆盖。
type Exporter struct {
	exportToFile bool
	path         string
	log          *logs.Wrapper
}

// NewExporter 创建文件导出器。
// 每次创建时，先将目录清空，避免历史文件干扰。
// path 默认是 "galileo/metrics"，不支持配置。
func NewExporter(exportToFile bool, path string, log *logs.Wrapper) *Exporter {
	f := &Exporter{
		exportToFile: exportToFile,
		path:         path,
		log:          log,
	}
	if err := os.RemoveAll(f.path); err != nil {
		f.log.Errorf("[galileo]os.RemoveAll|err=%v,path=%v", err, f.path)
	}
	return f
}

func (f *Exporter) Export(m interface{}) {
	if !f.exportToFile {
		return
	}
	err := os.MkdirAll(f.path, fs.ModePerm)
	if err != nil {
		f.log.Errorf("[galileo]os.MkdirAll|err=%v,path=%v", err, f.path)
		return
	}
	fileName := fmt.Sprintf("%s/%d.json", f.path, time.Now().UnixNano())
	bytes, err := json.Marshal(m)
	if err != nil {
		f.log.Errorf("[galileo]json.Marshal|err=%v,m=%v", err, m)
		return
	}
	err = ioutil.WriteFile(fileName, bytes, fs.ModePerm)
	if err != nil {
		f.log.Errorf("[galileo]ioutil.WriteFile|err=%v,fileName=%v,len=%d", err, fileName, len(bytes))
		return
	}
}

func (f *Exporter) ExportProfilesBatch(batch *model.ProfilesBatch) {
	if !f.exportToFile {
		return
	}
	dir := filepath.Join(f.path, strconv.FormatInt(batch.End, 10))
	if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
		f.log.Errorf("[galileo]os.MkdirAll|err=%v,path=%v", err, dir)
		return
	}
	for _, prof := range batch.Profiles {
		fileName := filepath.Join(dir, prof.Name)
		if err := ioutil.WriteFile(fileName, prof.Data, fs.ModePerm); err != nil {
			f.log.Errorf("[galileo]ioutil.WriteFile|err=%v,fileName=%v,len=%d", err, fileName, len(prof.Data))
			return
		}
	}
}
