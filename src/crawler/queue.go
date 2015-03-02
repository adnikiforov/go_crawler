package crawler

import (
	"encoding/csv"
	"os"
)

var (
	queueChannel chan string
)

func startQueue() {
	log.Info("Starting up queue...")
	csvfile, error := os.Open(CSV_FILE_LOCATION)
	if error != nil {
		log.Error("Can't parse CSV file, exit with error %s", error)
		os.Exit(1)
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
	reader.Comma = '\t'

	rawCSVdata, error := reader.ReadAll()
	if error != nil {
		log.Error("Can't parse CSV file, exit with error %s", error)
		os.Exit(1)
	}

	queueSize := len(rawCSVdata)
	queueChannel = make(chan string, queueSize)
	log.Info("Queue size %d", queueSize)
	for _, each := range rawCSVdata {
		queueChannel <- each[0]
	}
	close(queueChannel)
	log.Info("Queue started!")
}
