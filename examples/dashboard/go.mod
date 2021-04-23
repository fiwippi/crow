module github.com/fiwippi/examples/dashboard

go 1.16

require (
	github.com/deckarep/golang-set v1.7.1
	github.com/fiwippi/crow v0.0.0-20210422173432-5ab846c322fa
	github.com/influxdata/influxdb-client-go/v2 v2.2.3
	github.com/joho/godotenv v1.3.0 // indirect
)

replace github.com/fiwippi/crow/ => ../../
