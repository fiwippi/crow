package metrics

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
	"os"
	"time"
)

var client influxdb2.Client
var token, bucket, org string

func init() {
	godotenv.Load()

	token = os.Getenv("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN")
	bucket = os.Getenv("DOCKER_INFLUXDB_INIT_BUCKET")
	org = os.Getenv("DOCKER_INFLUXDB_INIT_ORG")

	client = influxdb2.NewClient("http://influxdb:8086", token)
}

func (s *stat) writeDB() {
	// get non-blocking write client
	writeAPI := client.WriteAPI(org, bucket)

	// create point using fluent style
	p := influxdb2.NewPointWithMeasurement("board_data").
		AddTag("board", s.Board).
		AddTag("duration", s.Duration).
		AddField(s.Name, s.Count).
		SetTime(time.Now())

	// write point asynchronously
	writeAPI.WritePoint(p)

	// Flush writes
	writeAPI.Flush()
}
