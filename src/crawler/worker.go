package crawler

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	EMAIL_REGEXP = regexp.MustCompile("href=[\\'|\"]mailto:([\\d\\w-\\.]+@[\\d\\w-\\.]+\\.[\\d\\w]+)")
	timeout      = time.Duration(5 * time.Second)
	transport    = http.Transport{Dial: dialTimeout}
	client       = http.Client{Transport: &transport}
)

func StartWorkers() {
	log.Info("Starting up workers...")
	for i := 0; i < WCOUNT; i++ {
		wg.Add(1)
		log.Debug("Starting up worker %d", i)
		go workerRoutine(QueueChannel)
	}
	log.Info("Started workers!")
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func workerRoutine(channel <-chan string) {
	WorkersCount++
	//	defer wg.Done()
	for {
		select {
		case domain := <-channel:
			if len(channel) > 0 {
				response, error := client.Get(string(strings.Join([]string{"http://", domain}, "")))
				if error != nil {
					StatData.domainError++
					StatData.overall++
					break
				}
				defer response.Body.Close()
				body, error := ioutil.ReadAll(response.Body)
				if error != nil {
					StatData.httpError++
					StatData.overall++
					break
				}
				applyRegexps(body, domain)

			} else {
				log.Info("Shutting down worker, (%d) workers left", WorkersCount)
				WorkersCount--
				if WorkersCount == 0 {
					log.Info("Shutting down...")
					os.Exit(0)
				}
				wg.Done()
				return
			}
		default:
			log.Info("Shutting down worker!")
			WorkersCount--
			wg.Done()
			return
		}
	}
}

func applyRegexps(body []byte, host string) {
	StatData.overall++
	matchCounter := 0
	for id, re := range Regexps {
		match := re.Match(body)
		if match {
			SaveParseResult(id, host)
			matchCounter++
		}
	}
	if matchCounter > 0 {
		StatData.matched++
	} else {
		StatData.noMatched++
	}
	tempMap := make(map[string]bool)
	emails := EMAIL_REGEXP.FindAllSubmatch(body, -1)
	for _, slice := range emails {
		str := CToGoString(slice[1])
		if !tempMap[str] {
			tempMap[str] = true
		}
	}
	if len(tempMap) > 0 {
		for email, _ := range tempMap {
			SaveEmailResult(host, email)
		}
	}
}

func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}

func updateQueueSize() {
	StatData.queueSize = len(QueueChannel)
}

func printStat() {
	log.Info("Processed %d domains, including %d matched, %d no matched, %d domain errors, %d http errors and %d domains left", StatData.overall, StatData.matched, StatData.noMatched, StatData.domainError, StatData.httpError, StatData.queueSize)
}
