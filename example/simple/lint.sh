go mod tidy

golangci-lint run -c .golangci.yml

go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
fieldalignment -fix ./
