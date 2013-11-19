package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/op/go-logging"
	"github.com/vmihailenco/redis"
	"os"
	"regexp"
)

var (
	path        string
	redisHost   string
	redisPasswd string
	logRegex    string
	redisDB     int64
	log         logging.Logger
)

func init() {
	flag.StringVar(&path, "l", "", "Path to the access log to watch")
	flag.StringVar(&redisHost, "h", "localhost:6379", "Hostname:port Redis")
	flag.StringVar(&redisPasswd, "p", "", "Redis password")
	flag.Int64Var(&redisDB, "d", -1, "Redis DB number to store the data")
	flag.StringVar(&logRegex, "r", "", "PCRE Regex to parse the log. Must return the remote IP on the first capture group")
}

func main() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "You need arguments")
		flag.Usage()
		os.Exit(10)
	}
	flag.Parse()

	// Seek to the end at the start
	seek := &tail.SeekInfo{
		Offset: 0,
		Whence: 2,
	}

	config := tail.Config{
		ReOpen:    true,
		MustExist: true,
		Follow:    true,
		Location:  seek,
	}

	t, err := tail.TailFile(path, config)
	if err != nil {
		log.Fatalf("Could not tail: %v\n", err)
	}

	client := redis.NewTCPClient(redisHost, redisPasswd, redisDB)
	multi, err := client.MultiClient()
	if err != nil {
		log.Fatalf("Failed to create multiclient: %v", err)
	}
	defer client.Close()
	log.Info("Connected to redis")

	r := regexp.MustCompile(logRegex)
	ipRegex := regexp.MustCompile(`\S+\.\S+\.\S+\.\S+`)

	for line := range t.Lines {
		if matches := r.FindStringSubmatch(line.Text); matches != nil {
			if !ipRegex.MatchString(matches[1]) {
				continue
			}
			if res := client.Ping(); res.Err() != nil {
				log.Warning("Redis not connected, %v, reconnecting", res.Val())
				client = redis.NewTCPClient(redisHost, redisPasswd, redisDB)
				multi, err = client.MultiClient()
				if err != nil {
					log.Fatalf("Failed to create multiclient: %v", err)
				}
			}
			_, err := multi.Exec(func() {
				multi.ZIncrBy("ipcount_5m", 1, matches[1])
				multi.ZIncrBy("ipcount_1h", 1, matches[1])
				multi.ZIncrBy("ipcount_12h", 1, matches[1])
				multi.ZIncrBy("ipcount_24h", 1, matches[1])
			})
			if err == redis.Nil {
				log.Warning("Failed to add to set: %v", err)
			}
		}
	}
}
