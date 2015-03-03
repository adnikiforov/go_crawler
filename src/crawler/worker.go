package crawler

import (
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	EMAIL_REGEXP = regexp.MustCompile("href=[\\'|\"]mailto:([\\d\\w-\\.]+@[\\d\\w-\\.]+\\.[\\d\\w]+)")
	timeout      = time.Duration(5 * time.Second)
	transport    = http.Transport{Dial: dialTimeout, DisableKeepAlives: true}
	client       = http.Client{Transport: &transport}
)

func startWorkers() {
	log.Info("Starting up workers...")
	for i := 0; i < WCOUNT; i++ {
		log.Debug("Starting up worker %d", i)
		go workerRoutine(queueChannel)
	}
	log.Info("Started workers!")
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func workerRoutine(channel <-chan string) {
	statWorkersCount++
	defer waitGroup.Done()
	for {
		//		select {
		//		case
		domain := <-channel
		//			if len(channel) > 0 {
		statOverall++
		response, error := client.Get(string(strings.Join([]string{"http://", domain}, "")))
		if error != nil {
			statDomainError++
			break
		}
		defer response.Body.Close()
		body, error := ioutil.ReadAll(response.Body)
		if error != nil {
			statHttpError++
			break
		}
		applyRegexps(body, domain)
		//			} else {
		//				statWorkersCount--
		//				log.Info("Shutting down worker, (%d) workers left", statWorkersCount)
		//				return
		//			}
		//		default:
		//			statWorkersCount--
		//			log.Info("Shutting down worker, (%d) workers left", statWorkersCount)
		//			return
		//		}
	}
}

func applyRegexps(body []byte, host string) {
	matchCounter := 0
	for id, re := range regexps {
		match := re.Match(body)
		if match {
			saveParseResult(id, host)
			matchCounter++
		}
	}
	if matchCounter > 0 {
		statMatched++
	} else {
		statNoMatched++
	}

	emails := EMAIL_REGEXP.FindAllSubmatch(body, -1)
	for _, slice := range emails {
		saveEmailResult(host, cToGoString(slice[1]))
	}
}

func cToGoString(c []byte) string {
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
	statQueueSize = len(queueChannel)
}

func printStat() {
	log.Info("Processed %d domains, including %d matched, %d no matched, %d domain errors, %d http errors and %d domains left",
		statOverall, statMatched, statNoMatched, statDomainError, statHttpError, statQueueSize)
}
