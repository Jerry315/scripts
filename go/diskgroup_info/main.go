package main

import (
	"dev/diskgroup_info/common"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
	"time"
)

const base_format = "2006-01-02 15:04:05"

func parseTime(t int64) string {
	return time.Unix(t, 0).Format(base_format)
}

func parseCapacity(c int) string {
	if (c / 1024 / 1024 / 1024) > 1024 {
		return fmt.Sprintf("%.2fTB", float64(c)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fGB", float64(c)/float64(1024*1024*1024))
	}
}

func groupInfo(discGroups *common.DiscGroups, discIdToTime map[int]int, cycleSum map[string]int, d bool) (queryStr string) {
	createTime := parseTime(int64(discGroups.Create_time))
	groupId := discGroups.Group_id
	groupType := discGroups.Group_type
	cycle := discGroups.Cycle
	onlines := discGroups.Onlines
	lockTime := discGroups.LockTime
	dispatcherId := discGroups.Dispatcher_ids
	totalUsed := 0
	totalLeft := 0
	disk_str := ""

	if _, ok := cycleSum[fmt.Sprintf("cycle_%d", cycle)]; ok {
		cycleSum[fmt.Sprintf("cycle_%d", cycle)]++
	} else {
		cycleSum[fmt.Sprintf("cycle_%d", cycle)] = 1
	}
	for i := 0; i < len(discGroups.Disc_infos)-1; i++ {
		for j := 0; j < len(discGroups.Disc_infos)-1-i; j++ {
			v1 := 0
			v2 := 0
			if v, ok := discIdToTime[discGroups.Disc_infos[j].DiscID]; ok {
				v1 = v
			}
			if v, ok := discIdToTime[discGroups.Disc_infos[j+1].DiscID]; ok {
				v2 = v
			}
			if v1 > v2 {
				continue
			} else {
				tmp := discGroups.Disc_infos[j]
				discGroups.Disc_infos[j] = discGroups.Disc_infos[j+1]
				discGroups.Disc_infos[j+1] = tmp
			}
		}

	}

	for i, disk := range discGroups.Disc_infos {
		totalUsed += disk.Used
		totalLeft += disk.Left
		if d {
			firstTime := "0"
			if v, ok := discIdToTime[disk.DiscID]; ok {
				firstTime = parseTime(int64(v))
			}

			sqd := i
			discID := disk.DiscID
			isOnline := disk.Is_online
			publicIp := disk.Public_ip
			localIp := disk.Local_ip
			port := disk.Port
			left := parseCapacity(disk.Left)
			tmp := fmt.Sprintf("\t[first_time: %s, seq: %d, discID: %d, is_online: %d, public_ip: %s, local_ip: %s, port: %d, left: %s]\n",
				firstTime, sqd, discID, isOnline, publicIp, localIp, port, left)
			if discGroups.Group_id == 0 {
				tmp = fmt.Sprintf("\t[first_time: %s, seq: %d, discID: %d, cycle: %d, is_online: %d, public_ip: %s, local_ip: %s, port: %d, left: %s]\n",
					firstTime, sqd, discID, disk.Cycle, isOnline, publicIp, localIp, port, left)
			}
			disk_str += tmp
		}

	}

	tmpStr := fmt.Sprintf("[%s] [Getdiskinfo; group_id: %d, group_type: %d, cycle: %d, onlines: %d, lockTime: %d, dispatcher_id: %v, total_used: %s, total_left: %s]\n",
		createTime, groupId, groupType, cycle, onlines, lockTime, dispatcherId, parseCapacity(totalUsed), parseCapacity(totalLeft))
	queryStr += tmpStr + disk_str
	return

}

func discIdToTime() map[int]int {
	conf := common.GetConf()
	firstTimeInfo := common.GetFirstTime(conf.Codisktrackerurl)
	discIdToTime := make(map[int]int)
	for _, timeInfo := range firstTimeInfo.Disc_first_time_infos {
		discIdToTime[timeInfo.Disc_id] = timeInfo.First_time
	}
	return discIdToTime
}

func createCycleStr(cycleSum map[string]int) (cycleStr string) {
	cycleStr = fmt.Sprintf("[%s] [", time.Now().Format(base_format))
	for k, v := range cycleSum {
		cycleStr += fmt.Sprintf("%s: %d, ", k, v)
	}
	cycleStr = strings.Trim(strings.Trim(cycleStr, " "), ",")
	cycleStr = cycleStr + "]"
	return
}

func summary(d bool) {
	conf := common.GetConf()
	groupData := common.GetGroupData(conf.Codisktrackerurl, conf.Limit)
	discIdToTime := discIdToTime()
	queryStr := ""
	cycleSum := make(map[string]int)
	for _, item := range groupData.Group_infos {
		queryStr += groupInfo(item, discIdToTime, cycleSum, d)

	}
	cycleStr := createCycleStr(cycleSum)
	queryStr += cycleStr
	fmt.Println(queryStr)
}

func queryGroupZeroCapacity() (capacityStr string) {
	conf := common.GetConf()
	groupData := common.GetGroupData(conf.Codisktrackerurl, conf.Limit)
	for _, item := range groupData.Group_infos {
		if item.Group_id == 0 {
			capacity := make(map[int]int)
			capacityStr = fmt.Sprintf("[%s] [", time.Now().Format(base_format))
			for _, disc := range item.Disc_infos {
				c := disc.Left + disc.Used
				if (c / (1024 * 1024 * 1024 * 1024)) > 4 {
					if _, ok := capacity[8]; ok {
						capacity[8]++
					} else {
						capacity[8] = 1
					}
				} else if 2 < (c/(1024*1024*1024*1024)) && (c/(1024*1024*1024*1024)) < 4 {
					if _, ok := capacity[4]; ok {
						capacity[4]++
					} else {
						capacity[4] = 1
					}
				} else {
					if _, ok := capacity[2]; ok {
						capacity[2]++
					} else {
						capacity[2] = 1
					}
				}
			}
			for k, v := range capacity {
				capacityStr += fmt.Sprintf("disk %dTB count: %d, ", k, v)
			}
			capacityStr = strings.Trim(strings.Trim(capacityStr, " "), ",")
			capacityStr = capacityStr + "]"
		}
	}
	return
}

func createQueryStr(discGroups *common.DiscGroups, discIdToTime map[int]int, cycleSum map[string]int, d bool, gid, cycle int) (queryStr string) {
	if gid == -1 {
		if cycle == -1 {
			queryStr = groupInfo(discGroups, discIdToTime, cycleSum, d)
		} else if discGroups.Cycle == cycle {
			queryStr = groupInfo(discGroups, discIdToTime, cycleSum, d)
		}
	} else if discGroups.Group_id == gid {
		if cycle == -1 {
			queryStr = groupInfo(discGroups, discIdToTime, cycleSum, d)
		} else if discGroups.Cycle == cycle {
			queryStr = groupInfo(discGroups, discIdToTime, cycleSum, d)
		}
	}
	return
}

func hostView(groupInfo *common.GroupInfos, host string, cycle, gid int) {
	var allPartition = []common.Partition{}

	for _, gi := range groupInfo.Group_infos {
		for _, dg := range gi.Disc_infos {
			p := common.Partition{}
			p.Cycle = dg.Cycle
			p.GroupId = gi.Group_id
			p.Port = dg.Port
			p.DiscID = dg.DiscID
			p.LocalIP = dg.Local_ip
			allPartition = append(allPartition, p)
		}
	}
	var partitionMap = map[string][]common.Partition{}
	for _, p := range allPartition {
		if host != "0.0.0.0" {
			if p.LocalIP != host {
				continue
			}
		}
		if cycle != -1 {
			if p.Cycle != cycle {
				continue
			}
		}
		if gid != -1 {
			if p.GroupId != gid {
				continue
			}
		}
		if _, ok := partitionMap[p.LocalIP]; ok {
			partitionMap[p.LocalIP] = append(partitionMap[p.LocalIP], p)
		} else {
			partitionMap[p.LocalIP] = []common.Partition{}
			partitionMap[p.LocalIP] = append(partitionMap[p.LocalIP], p)
		}
	}
	for _, ps := range partitionMap {
		for i := 0; i < len(ps); i++ {
			for j := 0; j < len(ps)-i-1; j++ {
				if ps[j].Port > ps[j+1].Port {
					tmp := ps[j]
					ps[j] = ps[j+1]
					ps[j+1] = tmp
				}
			}
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"LocalIP", "GroupID", "Port", "Cycle", "DiscID", "Allocation"})
	table.SetAutoMergeCells(false)
	table.SetRowLine(true)
	var data [][]string
	for _, ps := range partitionMap {
		gids := []string{}
		ports := []string{}
		cycles := []string{}
		dids := []string{}
		sts := []string{}
		localIP := ""
		for _, p := range ps {
			status := "Y"
			if p.GroupId == 0 {
				status = "N"
			}
			//data=append(data,[]string{p.LocalIP,strconv.Itoa(p.GroupId),strconv.Itoa(p.Port),strconv.Itoa(p.Cycle),strconv.Itoa(p.DiscID),status})
			gids = append(gids, strconv.Itoa(p.GroupId))
			ports = append(ports, strconv.Itoa(p.Port))
			cycles = append(cycles, strconv.Itoa(p.Cycle))
			dids = append(dids, strconv.Itoa(p.DiscID))
			sts = append(sts, status)
			localIP = p.LocalIP
		}
		data = append(data, []string{
			localIP,
			strings.Join(gids, "\n"),
			strings.Join(ports, "\n"),
			strings.Join(cycles, "\n"),
			strings.Join(dids, "\n"),
			strings.Join(sts, "\n"),
		})
	}
	table.AppendBulk(data)
	table.Render()
}

func queryGroupInfo(gid, cycle int, d bool) {
	conf := common.GetConf()
	groupData := common.GetGroupData(conf.Codisktrackerurl, conf.Limit)
	discIdToTime := discIdToTime()
	queryStr := ""
	cycleSum := make(map[string]int)
	for _, item := range groupData.Group_infos {
		queryStr += createQueryStr(item, discIdToTime, cycleSum, d, gid, cycle)
	}

	fmt.Println(strings.Trim(queryStr, "\n"))
}

func main() {
	app := cli.NewApp()
	app.Name = "show disk group info"
	app.Commands = []cli.Command{
		{
			Name: "summary",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "detail,d",
					Usage: "show detail",
				},
			},
			Action: func(c *cli.Context) {
				d := c.Bool("detail")
				summary(d)
			},
		},
		{
			Name: "group",
			Action: func(c *cli.Context) {
				groupId := c.Int("groupId")
				cycle := c.Int("cycle")
				d := c.Bool("detail")
				queryGroupInfo(groupId, cycle, d)
				if groupId == 0 {
					capacityStr := queryGroupZeroCapacity()
					if capacityStr != "" {
						fmt.Println(capacityStr)
					}
				}
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "groupId,g",
					Usage: "group id",
					Value: -1,
				},
				cli.IntFlag{
					Name:  "cycle,c",
					Usage: "group cycle",
					Value: -1,
				},
				cli.BoolFlag{
					Name:  "detail,d",
					Usage: "show detail",
				},
			},
		},
		{
			Name: "host",
			Action: func(c *cli.Context) {
				host := c.String("host")
				cycle := c.Int("cycle")
				gid := c.Int("groupId")
				conf := common.GetConf()
				groupData := common.GetGroupData(conf.Codisktrackerurl, conf.Limit)
				hostView(groupData, host, cycle, gid)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "host,H",
					Usage: "local ip",
					Value: "0.0.0.0",
				},
				cli.IntFlag{
					Name:  "cycle,c",
					Usage: "group cycle",
					Value: -1,
				},
				cli.IntFlag{
					Name:  "groupId,g",
					Usage: "group id",
					Value: -1,
				},
			},
		},
	}
	app.Run(os.Args)
}
