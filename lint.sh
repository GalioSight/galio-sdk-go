go mod tidy

golangci-lint run -c .golangci.yml

go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
fieldalignment -fix ./
fieldalignment -fix ./components/
fieldalignment -fix ./configs/
fieldalignment -fix ./errs/
fieldalignment -fix ./exporters/otp/metrics/
fieldalignment -fix ./helper/
fieldalignment -fix ./model/
fieldalignment -fix ./processors/omp/metrics/
fieldalignment -fix ./processors/omp/metrics/point/
cd ../../../proto/; make all
