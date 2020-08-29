package main

import (
	"dev/monitor_diskgroup/common"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"strconv"
)

func getGroupId(monitorUrl, groupUrl string) {
	monitorInfo := common.GetMonitorData(monitorUrl)
	groupInfo := common.GetGroupData(groupUrl)
	l := 0
	for _, item := range monitorInfo.Monitor_info {
		if item.Group_id == 0 {
			continue
		}
		l++
	}
	queryData := make(map[string][]map[string]string, l)
	queryData["data"] = make([]map[string]string, l)
	i := 0
	for _, item := range monitorInfo.Monitor_info {
		if item.Group_id == 0 {
			continue
		}
		for _, discGroup := range groupInfo.Group_infos {
			if discGroup.Group_id == item.Group_id {
				group := make(map[string]string)
				group["{#GROUPID}"] = strconv.Itoa(item.Group_id)
				group["{#MAX_UPLOAD_RATE}"] = strconv.Itoa(item.Max_upload_rate)
				group["{#LIMIT_CAPACITY}"] = strconv.Itoa(int(item.Size/1024/1024/1024/1024/2)+1) + "T"
				group["{#DISK_LIMIT}"] = strconv.Itoa(discGroup.StorageScheme[0]*20) + "G"
				queryData["data"][i] = group
				i++
			}
		}

	}
	b, err := json.Marshal(queryData)
	if err != nil {
		fmt.Println("json parse monitor info failed")
	}
	fmt.Println(string(b))
}

func getGroupCycle(url string) {
	var cycle []int

	monitorInfo := common.GetMonitorData(url)
	for _, item := range monitorInfo.Monitor_info {
		if len(cycle) > 0 {
			flag := false
			for _, c := range cycle {
				if item.Cycle == c {
					flag = true
					break
				}
			}
			if ! flag {
				cycle = append(cycle, item.Cycle)
			}
		} else {
			cycle = append(cycle, item.Cycle)
		}
	}
	cycleMap := make(map[string][]map[string]int)
	cycleMap["data"] = make([]map[string]int, len(cycle))
	for i, c := range cycle {
		tmp := make(map[string]int)
		tmp["{#CYCLE}"] = c
		cycleMap["data"][i] = tmp
	}
	b, err := json.Marshal(cycleMap)
	if err != nil {
		fmt.Println("json parse monitor info failed")
	}
	fmt.Println(string(b))
}

func getGroupCapacity(url string, gid int, percent bool) {
	var used, total int
	monitorInfo := common.GetMonitorData(url)
	for _, item := range monitorInfo.Monitor_info {
		if item.Group_id == gid {
			for _, disk := range item.Disk_infos {
				used = used + disk.Used
				total = total + disk.Size
			}
		}
	}
	if percent {
		fmt.Printf("%.2f\n", (float64(total)-float64(used))/float64(total)*100)
	} else {
		fmt.Println(total - used)
	}
}

func getCycleCapacity(url string, cycle int) (used, total int) {
	monitorInfo := common.GetMonitorData(url)
	for _, item := range monitorInfo.Monitor_info {
		if item.Cycle == cycle {
			for _, disk := range item.Disk_infos {
				used = used + disk.Used
				total = total + disk.Size
			}
		}
	}
	return
}

func getGroupRate(url string, gid int) {
	monitorInfo := common.GetMonitorData(url)
	for _, item := range monitorInfo.Monitor_info {
		if item.Group_id == gid {
			fmt.Println(item.Upload_rate)
		}
	}
}

func getZeroData(url string) map[int]map[int]map[string]int {
	zeroGroup := common.GetZeroData(url)
	zeroData := make(map[int]map[int]map[string]int)
	if zeroGroup.Group_infos == nil {
		zeroData[0] = make(map[int]map[string]int)
		zeroData[0][0] = make(map[string]int)
		zeroData[0][0]["count"] = 0
		zeroData[0][0]["left"] = 0
		return zeroData
	}
	data := zeroGroup.Group_infos[0]

	for _, item := range data.Disc_infos {
		if item.Is_online == 0 {
			continue
		}
		capacity := ((item.Used + item.Left) / (1024 * 1024 * 1024 * 1024))
		if capacity < 2 {
			capacity = 2
		} else if capacity < 4 {
			capacity = 4
		} else {
			capacity = 8
		}
		if _, ok := zeroData[item.Cycle]; ok {
			if _, ok := zeroData[item.Cycle][capacity]; ok {
				zeroData[item.Cycle][capacity]["count"]++
				zeroData[item.Cycle][capacity]["left"] = zeroData[item.Cycle][capacity]["left"] + item.Left
			} else {
				tmp := make(map[string]int)
				tmp["count"] = 1
				tmp["left"] = item.Left
				zeroData[item.Cycle][capacity] = tmp
			}
		} else {
			cycle := make(map[int]map[string]int)
			tmp := make(map[string]int)
			tmp["count"] = 1
			tmp["left"] = item.Left
			cycle[capacity] = tmp
			zeroData[item.Cycle] = cycle
		}
	}
	return zeroData
}

func queryZero(url string) {
	zeroData := getZeroData(url)
	l := 0
	for _, item := range zeroData {
		for range item {
			l++
		}
	}
	queryData := make(map[string][]map[string]string, l)
	queryData["data"] = make([]map[string]string, l)
	i := 0
	for c, item := range zeroData {
		for s, _ := range item {
			tmp := make(map[string]string)
			tmp["{#DISK_SPACE}"] = strconv.Itoa(s)
			tmp["{#STORAGE_CYCLE}"] = strconv.Itoa(c)
			queryData["data"][i] = tmp
			i++
		}

	}
	v, err := json.Marshal(queryData)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(v))

}

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger := common.InitLogger()
			logger.Error(fmt.Sprintf("%#v", err))
		}
	}()
	conf := common.GetConf()
	app := cli.NewApp()
	app.Name = "disk group monitor"
	app.Usage = " get disk group capacity or rate"
	app.Commands = []cli.Command{
		{
			Name:    "zero",
			Aliases: []string{"z"},

			Action: func(c *cli.Context) {
				cycle := c.Int("cycle")
				space := c.Int("space")
				zeroData := getZeroData(conf.GroupInfoUrl)
				if v, ok := zeroData[cycle][space]; ok {
					fmt.Println(v["left"])
				} else {
					fmt.Println(0)
				}
			},
			OnUsageError: nil,
			Subcommands: []cli.Command{
				{
					Name:  "query",
					Usage: "get zero group info",
					Action: func(c *cli.Context) error {
						queryZero(conf.GroupInfoUrl)
						return nil
					},
				},
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "cycle,c",
					Usage: "disk group cycle",
					Value: 0,
				},
				cli.IntFlag{
					Name:  "space,s",
					Usage: "disk group space",
					Value: 0,
				},
			},
		},
		{
			Name:  "capacity",
			Usage: "get disk group left capacity",
			Action: func(c *cli.Context) {
				groupId := c.Int("groupId")
				percent := c.Bool("percent")
				getGroupCapacity(conf.MonitorUrl, groupId, percent)
			},
			Subcommands: []cli.Command{
				{
					Name:  "query",
					Usage: "get disk group info",
					Action: func(c *cli.Context) error {
						getGroupId(conf.MonitorUrl, conf.GroupInfoUrl)
						return nil
					},
				},
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "groupId,g",
					Usage: "group id",
					Value: 0,
				},
				cli.BoolFlag{
					Name:  "percent,p",
					Usage: "display percent",
				},
			},
		},
		{
			Name:  "rate",
			Usage: "get disk group rate",
			Action: func(c *cli.Context) {
				groupId := c.Int("groupId")
				getGroupRate(conf.MonitorUrl, groupId)
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "groupId,g",
					Usage: "group id",
					Value: 0,
				},
			},
		},
		{
			Name:  "storage",
			Usage: "从磁盘存储周期的维度计算不同存储周期剩余容量",
			Action: func(c *cli.Context) {
				cycle := c.Int("cycle")
				percent := c.Bool("percent")
				total := c.Bool("total")
				left := c.Bool("left")
				u, t := getCycleCapacity(conf.MonitorUrl, cycle)
				if percent {
					fmt.Printf("%.4f\n", float64(t-u)/float64(t))
				}
				if total {
					fmt.Println(t)
				}
				if left {
					fmt.Println(t - u)
				}
			},
			Subcommands: []cli.Command{
				{
					Name:  "query",
					Usage: "get disk group cycle info",
					Action: func(c *cli.Context) error {
						getGroupCycle(conf.MonitorUrl)
						return nil
					},
				},
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "cycle,c",
					Usage: "disk group cycle",
					Value: 0,
				},
				cli.BoolFlag{
					Name:  "percent,p",
					Usage: "display left capacity percent",
				},
				cli.BoolFlag{
					Name:  "total,t",
					Usage: "total capacity",
				},
				cli.BoolFlag{
					Name:  "left,l",
					Usage: "left capacity",
				},
			},
		},
	}
	app.Run(os.Args)

}
