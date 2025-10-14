rm -f metrics.test;
go test -c -run TestNewMeterProvider;
./metrics.test "25s";
rm -f metrics.test;