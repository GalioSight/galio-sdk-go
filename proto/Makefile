# 需要安装protobuf

go_mod_path = $(shell go env GOMODCACHE)
gogo_sub_path = /github.com/gogo/protobuf@v1.3.2/
gogo_path = $(go_mod_path)$(gogo_sub_path)

all: gogo omp ocp otp omp-v3

omp:  model
	protoc -I=/usr/local/include/ -I=/usr/include/ -I=. -I=$(gogo_path) --gogofast_out=../model/ omp.proto

ocp:  model
	protoc -I=/usr/local/include/ -I=/usr/include/ -I=. -I=$(gogo_path) --gogofast_out=../model/ ocp.proto
	sh ./add_yaml_tag.sh ../model/ocp.pb.go

otp:  model
	protoc -I=/usr/local/include/ -I=/usr/include/ -I=. -I=$(gogo_path) --gogofast_out=../model/ otp.proto

clean:
	rm -rf ../model/omp.pb.go ../model/ocp.pb.go ../model/otp.pb.go

gogo:
	-go install github.com/gogo/protobuf/protoc-gen-gogofast@latest

model:
	mkdir -p ../model

model-v3:
	mkdir -p ../v3/model

omp-v3: model-v3
	protoc -I=/usr/local/include/ -I=/usr/include/ -I=. -I=$(gogo_path) --gogofast_out=../v3/model/ omp-v3.proto
