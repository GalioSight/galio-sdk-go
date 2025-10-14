set -eu
echo ">>> 开始运行 benchmark"

if ! which benchstat; then
  echo "error: 需要通过以下命令安装 benchstat:"
  echo "go install golang.org/x/perf/cmd/benchstat@latest"
  exit 1
fi

RESULT_DIR=$(pwd)/bench_result
TMP_FILE="tmp_file.txt"

mkdir -p $RESULT_DIR

go test -bench 'BenchmarkSimpleProfiler' -count=10 | tee $RESULT_DIR/simple_profiler.txt
go test -bench 'BenchmarkFastProfiler' -count=10 | tee $RESULT_DIR/fast_profiler.txt
go test -bench 'BenchmarkSimpleProfiler' -benchmem -count=10 -memprofilerate=1 | tee $RESULT_DIR/simple_profiler_mem.txt
go test -bench 'BenchmarkFastProfiler' -benchmem -count=10 -memprofilerate=1 | tee $RESULT_DIR/fast_profiler_mem.txt

cd $RESULT_DIR

sed 's/BenchmarkSimpleProfiler/BenchmarkDelta/g' simple_profiler.txt > cpu_simple.txt
awk '{print $1, $2, $3, $4, $5, $6}' cpu_simple.txt > $TMP_FILE && mv $TMP_FILE cpu_simple.txt
sed 's/BenchmarkSimpleProfiler/BenchmarkDelta/g' simple_profiler_mem.txt > mem_simple.txt
awk '{print $1, $2, $7, $8, $9, $10, $11, $12}' mem_simple.txt > $TMP_FILE && mv $TMP_FILE mem_simple.txt


sed 's/BenchmarkFastProfiler/BenchmarkDelta/g' fast_profiler.txt > cpu_fast.txt
awk '{print $1, $2, $3, $4, $5, $6}' cpu_fast.txt > $TMP_FILE && mv $TMP_FILE cpu_fast.txt
sed 's/BenchmarkFastProfiler/BenchmarkDelta/g' fast_profiler_mem.txt > mem_fast.txt
awk '{print $1, $2, $7, $8, $9, $10, $11, $12}' mem_fast.txt > $TMP_FILE && mv $TMP_FILE mem_fast.txt

# cpu
benchstat cpu_simple.txt cpu_fast.txt > cpu_benchstat.txt
# heap
benchstat mem_simple.txt mem_fast.txt > mem_benchstat.txt

# 合并 cpu 和 memory 的 benchstat 结果
cat cpu_benchstat.txt mem_benchstat.txt | tee benchstat.txt

# clean
rm cpu_simple.txt cpu_fast.txt mem_simple.txt mem_fast.txt
