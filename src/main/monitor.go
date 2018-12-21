package main

import (
	"fmt"
	"net/http"
)

// 查询监控信息的接口
func statsHandler(w http.ResponseWriter, r *http.Request) {
	_str := ""
	for _, v := range pBackendSvrs {
		_str += fmt.Sprintf("Server:%s FailTimes:%d isUp:%t\n", v.svrStr, v.failTimes, v.isUp)
	}
	fmt.Fprintf(w, "%s", _str)
}

func initStats() {
	pLog.Infof("Start monitor on addr %s", pConfig.Stats)

	go func() {
		http.HandleFunc("/stats", statsHandler)
		http.ListenAndServe(pConfig.Stats, nil)
	}()
}

// 健康状态信息的接口
func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "{\"status_code\":200}")
}

func initHealth() {
	pLog.Infof("Start health on addr %s", pConfig.Stats)
	http.HandleFunc("/health", healthHandler)
	http.ListenAndServe(pConfig.Stats, nil)
}
