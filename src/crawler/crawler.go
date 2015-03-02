package crawler

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"sync"
	"time"
)

var (
	DateStart    string
	wg           sync.WaitGroup
	WCOUNT       int   = 10
	log                = logging.MustGetLogger("crawler")
	format             = logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level} %{id:03x}%{color:reset} %{message}")
	StatData           = Stat{0, 0, 0, 0, 0, 0}
	WorkersCount int64 = 0
)

type Stat struct {
	matched     int64
	noMatched   int64
	domainError int64
	httpError   int64
	overall     int64
	queueSize   int
}

func fixDate() {
	t := time.Now()
	DateStart = string(fmt.Sprintf("%d.%d.%02d", t.Day(), t.Month(), t.Year()))
}

func initializeFlags() {
	flag.IntVar(&WCOUNT, "w", WCOUNT, "Workers count")
	flag.Parse()
}

func Run() {
	fixDate()
	initializeFlags()
	LoadRegexps()
	UpdateKeyList()
	log.Info("Starting up...")
	StartQueue()
	updateQueueSize()

	StartDb()
	StartWorkers()

	ticker := time.NewTicker(time.Millisecond * 10000)
	go func() {
		for t := range ticker.C {
			t.String() //Prevent compile "variable not used" error
			updateQueueSize()
			printStat()
		}
	}()

	wg.Wait()
	ticker.Stop()
}
