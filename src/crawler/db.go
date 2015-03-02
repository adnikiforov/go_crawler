package crawler

import (
	"github.com/fzzy/radix/redis"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	dbChannel chan DbObject
	regexps   map[int]*regexp.Regexp
)

type DbObject struct {
	Key   string
	Value string
}

func startDb() {
	dbChannel = make(chan DbObject)

	log.Info("Starting up db...")
	go saveRoutine(dbChannel)
	log.Info("Started db!")
}

func updateKeyList() {
	client, error := redis.Dial("tcp", "localhost:6379")
	defer client.Close()
	if error != nil {
		log.Error("Redis connection error %s", error)
		os.Exit(1)
	}

	for id := range regexps {
		key := strings.Join([]string{dateStart, strconv.Itoa(id)}, ":")
		error := client.Cmd("SADD", "domain_keys", key).Err
		if error != nil {
			log.Error("Redis update keylist error %s", error)
			os.Exit(1)
		}
	}
}

func saveParseResult(regexpId int, domain string) {
	dbChannel <- DbObject{
		Key:   strings.Join([]string{dateStart, strconv.Itoa(regexpId)}, ":"),
		Value: domain,
	}
}

func saveEmailResult(domain string, email string) {
	dbChannel <- DbObject{
		Key:   strings.Join([]string{domain, "email"}, ":"),
		Value: email,
	}
}

func loadRegexps() {
	client, error := redis.Dial("tcp", "localhost:6379")
	defer client.Close()
	if error != nil {
		log.Error("Redis connection error %s", error)
		os.Exit(1)
	}
	res, error := client.Cmd("HGETALL", "re").Hash()
	if error != nil {
		log.Error("Redis regexp loading error %s", error)
		os.Exit(1)
	}
	regexps = make(map[int]*regexp.Regexp)
	for key, value := range res {
		if strings.Split(value, ":")[2] == "active" {
			conv, err := strconv.Atoi(key)
			if err != nil {
				log.Error("Error on regexps reading, check syntax ~s", err)
			}
			regexps[conv] = regexp.MustCompile(strings.Split(value, ":")[1])
		}
	}
}

func saveRoutine(channel chan DbObject) {
	defer waitGroup.Done()
	client, error := redis.Dial("tcp", "localhost:6379")
	defer client.Close()
	if error != nil {
		log.Error("Redis connection error %s", error)
		return
	}
	for {
		var res DbObject
		res = <-channel
		client.Cmd("SADD", res.Key, res.Value)
	}
}
