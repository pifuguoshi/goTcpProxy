package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"stathat.com/c/consistent"
)

// BackendSvr Type
type BackendSvr struct {
	svrStr    string
	isUp      bool // is Up or Down
	failTimes int
}

var (
	pConsisthash *consistent.Consistent
	pBackendSvrs map[string]*BackendSvr
)

func initBackendSvrs(svrs []string) {
	pConsisthash = consistent.New()
	pBackendSvrs = make(map[string]*BackendSvr)

	for _, svr := range svrs {
		pConsisthash.Add(svr)
		pBackendSvrs[svr] = &BackendSvr{
			svrStr:    svr,
			isUp:      true,
			failTimes: 0,
		}
	}
	go checkBackendSvrs()
}

func getBackendSvr(conn net.Conn) (*BackendSvr, bool) {
	remoteAddr := conn.RemoteAddr().String()
	svr, _ := pConsisthash.Get(remoteAddr)
	bksvr, ok := pBackendSvrs[svr]
	return bksvr, ok
}

func callBackupServer(srv string) {
	client := &http.Client{}
	hostInfo := strings.Split(srv, ":")
	statusInfo := strings.Split(pConfig.Stats, ":")
	url := fmt.Sprintf("http://%s:%s/health", hostInfo[0], statusInfo[1])
	fmt.Println(url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	response, err := client.Do(request)
	pLog.Infof("call health %#v,%#v", response, err)
	backServer := pBackendSvrs[srv]
	if response != nil && response.StatusCode != 200 || err != nil {
		backServer.failTimes += 1
	} else {
		for _, v := range pConsisthash.Members() {
			if v == srv {
				return
			}
		}
		pConsisthash.Add(srv)
		backServer.failTimes = 0
		backServer.isUp = true
	}
}
func checkBackendSvrs() {
	// scheduler every 10 seconds
	rand.Seed(time.Now().UnixNano())
	t := time.Tick(time.Duration(10)*time.Second + time.Duration(rand.Intn(100))*time.Millisecond*100)

	for _ = range t {
		for _, v := range pBackendSvrs {
			callBackupServer(v.svrStr)
		}
		for _, v := range pBackendSvrs {
			if v.failTimes >= pConfig.FailOver && v.isUp == true {
				v.isUp = false
				pConsisthash.Remove(v.svrStr)
			}
		}

	}
}
