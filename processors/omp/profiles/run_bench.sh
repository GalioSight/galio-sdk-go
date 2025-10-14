set -eu
echo ">>> 开始运行 benchmark (ETA: 40min)"
BENCH_CMD="go test -benchtime 60s -count 6 -timeout 24h -run ^$ -bench . ."

if ! which benchstat; then
  echo "error: 需要通过以下命令安装 benchstat:"
  echo "go install golang.org/x/perf/cmd/benchstat@latest"
  exit 1
fi

RESULT_DIR=$(pwd)/bench_result

mkdir -p $RESULT_DIR

$BENCH_CMD | tee $RESULT_DIR/baseline.txt

env \
  ENABLE_PROFILE=true \
  $BENCH_CMD | tee $RESULT_DIR/enable_profiles.txt

cd $RESULT_DIR

# Format enable_profiles.txt
sed -i ':a;N;$!ba;s/runtime: cannot set cpu profile rate until previous profile has finished.\n//g' enable_profiles.txt

benchstat baseline.txt enable_profiles.txt | tee benchstat.txt
