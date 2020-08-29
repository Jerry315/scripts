package main

import (
	"dev/mail_api/common"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func sendReport(w http.ResponseWriter, r *http.Request) {
	conf := common.GetConf()
	logger := common.InitLogger()
	text := r.FormValue("text")
	err := common.SendMail(conf, text)
	if err != nil {
		logger.Error(fmt.Sprintf("send mail fail. %v", err))
	}

}

func routeFunc(conf common.Config) {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v1/notice", sendReport)
	http.ListenAndServe(conf.Bind, router)
}

func main() {
	conf := common.GetConf()
	routeFunc(conf)
}
