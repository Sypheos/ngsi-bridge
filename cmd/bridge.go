package main

import (
	"flag"
	"io/ioutil"
	"ngsi-bridge"
	"os"

	"github.com/TheThingsNetwork/go-utils/log/apex"
	"gopkg.in/yaml.v2"
)

var (
	bridge     string
	appID      string
	appKey     string
	account    string
	discovery  string
	httpPort   int
	mapperFile string
	brokerURL  string
	httpMethod string
)

func init() {
	flag.StringVar(&bridge, "type", "ttn", "Bridge type")
	flag.StringVar(&brokerURL, "broker", "http://localhost:1026/ngsi10/updateContext", "Fiware broker url")
	flag.StringVar(&mapperFile, "mapper", "./mapper.yml", "message mapper configuratio")
	// TTN
	flag.StringVar(&appID, "appID", "", "TTN application ID")
	flag.StringVar(&appKey, "appKey", "", "TTN application Key")
	flag.StringVar(&account, "account", "https://account.thethings.network", "TTN account server")
	flag.StringVar(&discovery, "discovery", "discovery.thethings.network:1900", "TTN discovery server")

	// HTTP
	flag.IntVar(&httpPort, "port", 8080, "Http server port")
	flag.StringVar(&httpMethod, "method", "POST", "Http server port")
}

func main() {
	flag.Parse()
	b := bridges.NewHttpBridge(httpPort)
	aLog := apex.Stdout()
	aLog.Level = apex.DebugLevel

	mapperConf, err := getMapperSchema(mapperFile)
	if err != nil {
		aLog.Error(err.Error())
		os.Exit(-1)
	}
	b.Prepare(aLog, mapperConf, brokerURL, httpMethod)
	if err = b.Open(); err != nil {
		aLog.Error(err.Error())
		os.Exit(-1)
	}
	return
}

func getMapperSchema(filename string) (map[string]bridges.Schema, error) {
	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	m := map[string]bridges.Schema{}
	err = yaml.Unmarshal(buff, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
