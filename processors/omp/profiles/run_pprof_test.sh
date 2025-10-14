set -eu
echo ">>> 开始运行 sonic 与 pprof 的冲突测试 (ETA: 10min)"
go test  -v -timeout 660s -run TestSonicPprofConflict ./pprof_test.go