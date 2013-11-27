package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
	"github.com/vmihailenco/redis"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

var (
	redisHost   string
	redisPasswd string
	redisDB     int64
	sharePath   string
	serverPort  string
	redisConn   *redis.Client
	log         = logging.MustGetLogger("ipcounttop")
	keyMap      = map[int]string{
		1: "ipcount_5m",
		2: "ipcount_1h",
		3: "ipcount_12h",
		4: "ipcount_24h",
	}
)

func init() {
	flag.StringVar(&redisHost, "h", "localhost:6379", "Hostname:port Redis")
	flag.StringVar(&redisPasswd, "p", "", "Redis password")
	flag.Int64Var(&redisDB, "d", -1, "Redis DB number to store the data")
	flag.StringVar(&sharePath, "s", "./src/github.com/coredump/ipcount/ipcounttop/webapp", "Path to webapp files dir")
	flag.StringVar(&serverPort, "l", ":8888", "Port to use for webserver :<port number> format")
}

type TopController struct{}

func (i *TopController) ReadMany(ctx context.Context) (err error) {

	type top struct {
		Id   int64
		Name string
	}
	topList := []top{
		top{Id: 1, Name: "5 minutes"},
		top{Id: 2, Name: "1 hour"},
		top{Id: 3, Name: "12 hours"},
		top{Id: 4, Name: "24 hours"},
	}
	return goweb.API.RespondWithData(ctx, topList)
}

func (t *TopController) Read(id string, ctx context.Context) (err error) {
	idi, err := strconv.Atoi(id)
	if err != nil {
		return goweb.API.RespondWithError(ctx, 500, err.Error())
	}
	key, exist := keyMap[idi]
	if !exist {
		return goweb.API.RespondWithError(ctx, 404, "Scoreboard not found")
	}
	redisConn, err := redisConnect(redisHost, redisPasswd, redisDB)
	if err != nil {
		return goweb.API.RespondWithError(ctx, 500, err.Error())
	}
	defer redisConn.Close()

	log.Debug("Getting data from %s", key)
	topHundredList := redisConn.ZRevRange(key, "0", "100")
	if topHundredList.Err() != nil {
		return goweb.API.RespondWithError(ctx, 500, topHundredList.Err().Error())
	}
	topHundredMap := redisConn.ZRevRangeWithScoresMap(key, "0", "100")
	if topHundredMap.Err() != nil {
		return goweb.API.RespondWithError(ctx, 500, topHundredMap.Err().Error())
	}
	var ret [][]string
	for _, v := range topHundredList.Val() {
		ret = append(ret, []string{v, fmt.Sprintf("%.0f", topHundredMap.Val()[v])})
	}
	return goweb.API.RespondWithData(ctx, ret)
}

func getWhois(c context.Context) error {
	ip := c.PathParams().Get("ip").Str()
	cmd := exec.Command("/usr/bin/whois", ip)
	log.Debug("%v", cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return goweb.API.RespondWithError(c, 500, err.Error())
	}
	return goweb.API.RespondWithData(c, out.String())
}

func redisConnect(addr, password string, db int64) (client *redis.Client, err error) {
	client = redis.NewTCPClient(addr, password, db)
	err = client.Ping().Err()
	return
}

func main() {
	flag.Parse()

	topController := new(TopController)
	goweb.MapController("/ipcount/top/", topController)
	goweb.MapStatic("/ipcount/a", sharePath)
	goweb.Map("/ipcount/whois/{ip}", getWhois)

	goweb.MapAfter(func(c context.Context) error {
		log.Info("%s %s", c.HttpRequest().RemoteAddr, c.HttpRequest().RequestURI)
		return nil
	})

	goweb.Map("/", func(c context.Context) error {
		return goweb.Respond.WithPermanentRedirect(c, "/ipcount/a/")
	})

	goweb.Map(func(c context.Context) error {
		return goweb.API.Respond(c, 404, nil, []string{"File not found"})
	})

	s := &http.Server{
		Addr:           serverPort,
		Handler:        goweb.DefaultHttpHandler(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Info("Starting server...")
	log.Fatalf("Error in ListenAndServe: %s", s.ListenAndServe())
}
