package main

import (
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/vmihailenco/redis"
	"os"
	"strconv"
	"time"
)

var (
	redisHost   string
	redisPasswd string
	redisDB     int64
	debug       = false
	log         = logging.MustGetLogger("ipcountclean")
)

func init() {
	flag.StringVar(&redisHost, "h", "localhost:6379", "Hostname:port Redis")
	flag.StringVar(&redisPasswd, "p", "", "Redis password")
	flag.Int64Var(&redisDB, "d", -1, "Redis DB number to store the data")
	flag.BoolVar(&debug, "debug", false, "Show a lot of probably useless information")
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "You need arguments")
		flag.Usage()
		os.Exit(10)
	}
	flag.Parse()

	if !debug {
		logging.SetLevel(logging.INFO, "ipcountclean")
	}

	client := redis.NewTCPClient(redisHost, redisPasswd, redisDB)
	defer client.Close()

	now := time.Now().Unix()
	five := now - int64((time.Duration(5) * time.Minute).Seconds())
	oneH := now - int64((time.Duration(1) * time.Hour).Seconds())
	twelve := now - int64((time.Duration(12) * time.Hour).Seconds())
	twentyF := now - int64((time.Duration(24) * time.Hour).Seconds())

	log.Debug("%d\n%d\n%d\n%d\n%d\n", now, five, oneH, twelve, twentyF)

	ipHashFive := client.HGetAllMap("ipcount_h5m")
	if ipHashFive.Err() != nil {
		log.Warning("Failed to get 5m hashes %v", ipHashFive.Err())
	}
	ipHashOneH := client.HGetAllMap("ipcount_h1h")
	if ipHashOneH.Err() != nil {
		log.Warning("Failed to get 1h hashes %v", ipHashOneH.Err())
	}
	ipHashTwelve := client.HGetAllMap("ipcount_h12h")
	if ipHashTwelve.Err() != nil {
		log.Warning("Failed to get 12h hashes %v", ipHashTwelve.Err())
	}
	ipHashTwenty := client.HGetAllMap("ipcount_h24h")
	if ipHashTwenty.Err() != nil {
		log.Warning("Failed to get 24h hashes %v", ipHashTwenty.Err())
	}

	// clear 5m
	keys, err := findKeys(ipHashFive.Val(), five)
	if err != nil {
		log.Warning("Failed to get keys: %v", err)
	}
	err = deleteKey(keys, "ipcount_h5m", "ipcount_5m", client)
	if err != nil {
		log.Warning("Failed to delete keys: %v", err)
	}

	// clear 1h
	keys, err = findKeys(ipHashOneH.Val(), oneH)
	if err != nil {
		log.Warning("Failed to get keys: %v", err)
	}
	err = deleteKey(keys, "ipcount_h1h", "ipcount_1h", client)
	if err != nil {
		log.Warning("Failed to delete keys: %v", err)
	}

	// clear 12h
	keys, err = findKeys(ipHashTwelve.Val(), twelve)
	if err != nil {
		log.Warning("Failed to get keys: %v", err)
	}
	err = deleteKey(keys, "ipcount_h12h", "ipcount_12h", client)
	if err != nil {
		log.Warning("Failed to delete keys: %v", err)
	}

	// clear 24h
	keys, err = findKeys(ipHashTwenty.Val(), twentyF)
	if err != nil {
		log.Warning("Failed to get keys: %v", err)
	}
	err = deleteKey(keys, "ipcount_h24h", "ipcount_24h", client)
	if err != nil {
		log.Warning("Failed to delete keys: %v", err)
	}

	log.Debug("Now: %d", now)
}

func findKeys(hash map[string]string, t int64) (keys []string, err error) {
	log.Debug("Map contains %v entries", len(hash))
	log.Debug("Finding keys")
	for k, v := range hash {
		ts, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return keys, fmt.Errorf("Failed to convert to int: %v", err)
		}
		if ts < t {
			keys = append(keys, k)
		}
	}
	log.Debug("Done finding keys")
	return
}

func deleteKey(keys []string, set, sset string, client *redis.Client) (err error) {
	if len(keys) == 0 {
		log.Debug("Nothing to delete")
		return nil
	}
	log.Debug("Deleting %d keys for %s", len(keys), sset)
	ret := client.ZRem(sset, keys...)
	if ret.Err() != nil {
		return fmt.Errorf("Failed to remove keys: %v", ret.Err())
	}
	ret = client.HDel(set, keys...)
	if ret.Err() != nil {
		return fmt.Errorf("Failed to remove keys: %v", ret.Err())
	}
	log.Debug("All done for %s", sset)
	return
}
