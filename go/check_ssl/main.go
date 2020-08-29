package main

import (
	"dev/check_ssl/cert"
	"dev/check_ssl/common"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"strings"
	"time"
)

func nsParse(domain, lvsUrl string, log *zap.Logger) (hosts []string, err error) {
	ns, err := net.LookupHost(domain)
	if err != nil {
		log.Error(fmt.Sprintf("lookup %s failed", domain))
		return
	}
	for _, n := range ns {
		newUrl := lvsUrl + n
		req, err := http.NewRequest("GET", newUrl, nil)
		if err != nil {
			log.Error(fmt.Sprintf("create request %s failed. %v", newUrl, err))
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error(fmt.Sprintf("execute request %s failed. %v", newUrl, err))
			continue
		}
		if resp.StatusCode == 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			for _, host := range strings.Split(string(body), ";") {
				hosts = append(hosts, host)
			}
		}
	}
	return
}

func Query(conf common.Config, log *zap.Logger) {
	result := make(map[string][]map[string]string)
	count := 0
	for _, domain := range conf.Domain {
		tmp1 := []string{}
		tmp2 := []string{}
		hosts, _ := nsParse(domain.Url, conf.LvsUrl, log)
		for _, host := range hosts {
			tmp1 = append(tmp1, host)
		}
		for _, host := range domain.Hosts {
			tmp1 = append(tmp1, host)
		}
		for i := range tmp1 {
			flag := true
			for j := range tmp2 {
				if tmp1[i] == tmp2[j] {
					flag = false
					break
				}
			}
			if flag {
				tmp2 = append(tmp2, tmp1[i])
			}
		}
		count += len(tmp2)

	}
	i := 0
	result["data"] = make([]map[string]string, count)
	for _, domain := range conf.Domain {
		tmp1 := []string{}
		tmp2 := []string{}
		hosts, _ := nsParse(domain.Url, conf.LvsUrl, log)
		for _, host := range hosts {
			tmp1 = append(tmp1, host)
		}

		for _, host := range domain.Hosts {
			tmp1 = append(tmp1, host)
		}
		for j := range tmp1 {
			flag := true
			for k := range tmp2 {
				if tmp1[j] == tmp2[k] {
					flag = false
					break
				}
			}
			if flag {
				tmp2 = append(tmp2, tmp1[j])
			}
		}
		for _, host := range tmp2 {
			result["data"][i] = make(map[string]string)
			result["data"][i]["{#HOST}"] = host
			result["data"][i]["{#URL}"] = domain.Url
			i++
		}

	}
	b, err := json.Marshal(result)
	if err != nil {
		log.Error("json parse result failed")
	}
	fmt.Println(string(b))
}

func main() {
	conf := common.GetConf()
	log := common.InitLogger()
	app := cli.NewApp()
	app.Name = "check_ssl"
	app.Commands = []cli.Command{
		{
			Name:        "monitor",
			Aliases:     []string{"m"},
			Usage:       "monitor domain ssl certificate info",
			Description: "monitor domain ssl certificate info",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "domain,d",
					Value: "",
					Usage: "domain url",
				},
				cli.StringFlag{
					Name:  "hostport,H",
					Value: "",
					Usage: "host and port ,x.x.x.x:port",
				},
			},
			Action: func(c *cli.Context) {
				doamin := c.String("domain")
				hostport := c.String("hostport")
				if doamin != "" && hostport != "" {
					result := cert.NewCert(hostport, doamin, log)
					if result.Error != "" {
						log.Error(result.Error)
						os.Exit(1)
					}
					expire_date := result.NotAfer
					expire_date = strings.Join(strings.Split(expire_date, " ")[:2], " ")
					t, err := time.Parse("2006-01-02 15:04:05", expire_date)
					if err != nil {
						log.Error(fmt.Sprintf("#%v", err))
					}
					h, _ := time.ParseDuration("8h")
					t = t.Add(h)
					p := t.Sub(time.Now())
					fmt.Println(int(p.Hours() / 24))
				} else {
					log.Error("cert.NewCert need two arguments.")
				}
			},
		},
		{
			Name:        "query",
			Aliases:     []string{"q"},
			Usage:       "monitor domain ssl certificate info",
			Description: "This is get doamin and hosts record",
			Action: func(c *cli.Context) {
				Query(conf, log)
			},
		},
	}

	app.Run(os.Args)
}
