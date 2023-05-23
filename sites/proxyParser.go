package sites

import (
	"math/rand"
	"strings"
)

func GetProxy() string {
	proxes := GetProxyList()
	proxyToParse := proxes[rand.Intn(len(proxes))]
	splitter := strings.Split(proxyToParse, ":")
	url := splitter[0]
	port := splitter[1]
	user := splitter[2]
	pass := splitter[3]
	proxy := "http://" + user + ":" + pass + "@" + url + ":" + port
	return proxy
}

func GetProxyList() [1]string {
	proxies := [...]string{ // We only used 4 proxies as they weren't static, so we just had them defined here, if you have more, best bet is to load from a .txt or other files
		"127.0.0.1:8080:ralph:lauren",
	}
	return proxies
}
