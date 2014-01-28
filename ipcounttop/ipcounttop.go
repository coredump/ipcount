package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/nranchev/go-libGeoIP"
	"github.com/op/go-logging"
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
	"github.com/vmihailenco/redis"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	log       = logging.MustGetLogger("ipcounttop")
	geoIPData string
	keyMap    = map[int]string{
		1: "ipcount_5m",
		2: "ipcount_1h",
		3: "ipcount_12h",
		4: "ipcount_24h",
	}
	redisHost   string
	redisPasswd string
	redisDB     int64
	sharePath   string
	serverPort  string
	redisConn   *redis.Client
)

func init() {
	flag.Int64Var(&redisDB, "d", -1, "Redis DB number to store the data")
	flag.StringVar(&geoIPData, "g", "/usr/share/GeoIP/GeoIP.dat", "Port to use for webserver :<port number> format")
	flag.StringVar(&redisHost, "h", "localhost:6379", "Hostname:port Redis")
	flag.StringVar(&serverPort, "l", ":8888", "Port to use for webserver :<port number> format")
	flag.StringVar(&redisPasswd, "p", "", "Redis password")
	flag.StringVar(&sharePath, "s", "./src/github.com/coredump/ipcount/ipcounttop/webapp", "Path to webapp files dir")
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
	idx, err := strconv.Atoi(id)
	if err != nil {
		return goweb.API.RespondWithError(ctx, 500, err.Error())
	}
	_, exist := keyMap[idx]
	if !exist {
		return goweb.API.RespondWithError(ctx, 404, "Scoreboard not found")
	}
	topHundredMap, topHundredList, err := getTopData(idx)
	if err != nil {
		return goweb.API.RespondWithError(ctx, 500, err.Error())
	}
	var ret [][]string
	for _, v := range topHundredList.Val() {
		countryData, err := getCountry(v)
		if err != nil {
			log.Info("Failed getting geo data for %s", v)
		}
		ret = append(ret, []string{v, fmt.Sprintf("%.0f", topHundredMap.Val()[v]), countryData[0], countryData[1]})
	}
	return goweb.API.RespondWithData(ctx, ret)
}

func getTopData(idx int) (topHundredMap *redis.StringFloatMapReq, topHundredList *redis.StringSliceReq, err error) {
	redisConn, err := redisConnect(redisHost, redisPasswd, redisDB)
	if err != nil {
		return
	}
	defer redisConn.Close()
	key, _ := keyMap[idx]
	log.Debug("Getting data from %s", key)
	topHundredList = redisConn.ZRevRange(key, "0", "100")
	if topHundredList.Err() != nil {
		return
	}
	topHundredMap = redisConn.ZRevRangeWithScoresMap(key, "0", "100")
	if topHundredMap.Err() != nil {
		return
	}
	return
}

func getWhois(ctx context.Context) error {
	r := regexp.MustCompile(`.+/(\S+\.\S+\.\S+\.\S+)$`)
	m := r.FindStringSubmatch(ctx.Path().RawPath)
	if len(m) != 2 {
		log.Info("IP not matched on path: %s", ctx.Path().RawPath)
		return goweb.API.RespondWithError(ctx, 500, "This IP didn't match")
	}
	ip := m[1]
	cmd := exec.Command("/usr/bin/whois", ip)
	log.Debug("%v", cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return goweb.API.RespondWithError(ctx, 500, err.Error())
	}
	return goweb.API.RespondWithData(ctx, out.String())
}

func getCountry(ip string) ([]string, error) {
	log.Debug("Getting country for %s", ip)
	geo, err := libgeo.Load(geoIPData)
	if err != nil {
		log.Debug("Failed to open location data")
		return []string{}, err
	}
	loc := geo.GetLocationByIP(ip)
	log.Debug("loc: %v", loc)
	if loc != nil {
		return []string{loc.CountryCode, loc.CountryName}, nil
	}
	return []string{"--", "N/A"}, nil
}

func getGeo(ctx context.Context) error {
	id := ctx.PathParams().Get("id").MustStr()
	geo, err := libgeo.Load(geoIPData)
	type geometry struct {
		Type        string    `json:"type"`
		Coordinates []float32 `json:"coordinates"`
	}
	type properties struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	type geojson struct {
		Type       string     `json:"type"`
		Geometry   geometry   `json:"geometry"`
		Properties properties `json:"properties"`
	}
	var data []geojson
	if err != nil {
		return goweb.API.RespondWithError(ctx, 500, err.Error())
	}
	idx, err := strconv.Atoi(id)
	if err != nil {
		return goweb.API.RespondWithError(ctx, 500, err.Error())
	}
	_, exist := keyMap[idx]
	if !exist {
		return goweb.API.RespondWithError(ctx, 404, "Scoreboard not found")
	}
	_, topHundredList, err := getTopData(idx)
	for _, ip := range topHundredList.Val() {
		loc := geo.GetLocationByIP(ip)
		if loc == nil {
			continue
		}
		json := geojson{"Feature",
			geometry{"Point", []float32{loc.Longitude, loc.Latitude}},
			properties{loc.City, fmt.Sprintf("Access from %s, %s", loc.City, loc.CountryName)}}
		data = append(data, json)
	}
	return goweb.API.RespondWithData(ctx, data)
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
	goweb.Map("/ipcount/mapdata/{id}", getGeo)

	goweb.MapAfter(func(ctx context.Context) error {
		log.Info("%s %s", ctx.HttpRequest().RemoteAddr, ctx.HttpRequest().RequestURI)
		return nil
	})

	goweb.Map("/", func(ctx context.Context) error {
		return goweb.Respond.WithPermanentRedirect(ctx, "/ipcount/a/")
	})

	goweb.Map(func(ctx context.Context) error {
		return goweb.API.Respond(ctx, 404, nil, []string{"File not found"})
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
