package crawler

import (
	"encoding/csv"
	"os"
)

var (
	QueueChannel chan string
)

func StartQueue() {
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
	QueueChannel = make(chan string, queueSize)
	log.Info("Queue size %d", queueSize)
	for _, each := range rawCSVdata {
		QueueChannel <- each[0]
	}
	close(QueueChannel)
	log.Info("Queue started!")
}
