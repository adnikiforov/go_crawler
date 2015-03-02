package crawler

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

var (
	dateStart string
	waitGroup sync.WaitGroup
	log       = logging.MustGetLogger("crawler")
)

var (
	statMatched      int64 = 0
	statNoMatched    int64 = 0
	statDomainError  int64 = 0
	statHttpError    int64 = 0
	statOverall      int64 = 0
	statWorkersCount int64 = 0
	statQueueSize    int   = 0
)

var (
	WCOUNT            int    = 100
	CSV_FILE_LOCATION string = "./domains.csv"
)

func Run() {
	//	Run profiler
	go func() { http.ListenAndServe("localhost:6060", nil) }()

	//	Fix startup date
	t := time.Now()
	dateStart = string(fmt.Sprintf("%d.%d.%02d", t.Day(), t.Month(), t.Year()))

	flag.IntVar(&WCOUNT, "w", WCOUNT, "Workers count")
	flag.StringVar(&CSV_FILE_LOCATION, "f", CSV_FILE_LOCATION, "Source file location")
	flag.Parse()

	loadRegexps()
	updateKeyList()

	log.Info("Starting up...")
	startQueue()
	updateQueueSize()
	startDb()

	waitGroup.Add(WCOUNT + 1)
	startWorkers()

	ticker := time.NewTicker(time.Millisecond * 10000)
	go func() {
		for t := range ticker.C {
			t.String() //Prevent compile "variable not used" error
			updateQueueSize()
			printStat()
		}
	}()

	waitGroup.Wait()
	ticker.Stop()
}
