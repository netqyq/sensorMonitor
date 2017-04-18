dst="./bin"
mkdir -p $dst


go build -o $dst/coordinator src/distributed/coordinator/exec/main.go
go build -o $dst/datamanager src/distributed/datamanager/exec/main.go
go build -o $dst/sensor src/distributed/sensor/sensor.go
go build -o $dst/web src/distributed/web/main.go
