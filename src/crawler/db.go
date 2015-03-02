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
	Regexps   map[int]*regexp.Regexp
)

type DbObject struct {
	Key   string
	Value string
}

func StartDb() {
	dbChannel = make(chan DbObject)
	wg.Add(1)
	log.Info("Starting up db...")
	go saveRoutine(dbChannel)
	log.Info("Started db!")
}

func UpdateKeyList() {
	client, error := redis.Dial("tcp", "localhost:6379")
	defer client.Close()
	if error != nil {
		log.Error("Redis connection error %s", error)
		os.Exit(1)
	}

	keys := make([]int, len(Regexps))
	i := 0
	for k := range Regexps {
		keys[i] = k
		i += 1
	}

	for _, id := range keys {
		key := strings.Join([]string{DateStart, strconv.Itoa(id)}, ":")
		error := client.Cmd("SADD", "domain_keys", key).Err
		if error != nil {
			log.Error("Redis update keylist error %s", error)
			os.Exit(1)
		}
	}
}

func SaveParseResult(regexpId int, domain string) {
	dbChannel <- DbObject{
		Key:   strings.Join([]string{DateStart, strconv.Itoa(regexpId)}, ":"),
		Value: domain,
	}
}

func SaveEmailResult(domain string, email string) {
	dbChannel <- DbObject{
		Key:   strings.Join([]string{domain, "email"}, ":"),
		Value: email,
	}
}

func LoadRegexps() {
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
	Regexps = make(map[int]*regexp.Regexp)
	for key, value := range res {
		if strings.Split(value, ":")[2] == "active" {
			conv, err := strconv.Atoi(key)
			if err != nil {
				log.Error("Error on regexps reading, check syntax ~s", err)
			}
			Regexps[conv] = regexp.MustCompile(strings.Split(value, ":")[1])
		}
	}
}

func saveRoutine(channel chan DbObject) {
	defer wg.Done()
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
