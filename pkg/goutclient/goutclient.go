package goutclient

import (
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	"strings"
)

var goutClient = gout.NewWithOpt(gout.WithInsecureSkipVerify())

func GetClient() *gout.Client {
	return goutClient
}

func GetClientGet(uri string, proxy string) *dataflow.DataFlow {
	if len(proxy) > 0 {
		if strings.Contains(proxy, "socks5://") {
			proxy = strings.Replace(proxy, "socks5://", "", 1)
			return goutClient.GET(uri).SetSOCKS5(proxy)
		} else {
			return goutClient.GET(uri).SetProxy(proxy)
		}
	} else {
		return goutClient.GET(uri)
	}
}

func GetClientPost(uri string, proxy string) *dataflow.DataFlow {
	if len(proxy) > 0 {
		if strings.Contains(proxy, "socks5://") {
			proxy = strings.Replace(proxy, "socks5://", "", 1)
			return goutClient.POST(uri).SetSOCKS5(proxy)
		} else {
			return goutClient.POST(uri).SetProxy(proxy)
		}
	} else {
		return goutClient.POST(uri)
	}
}
