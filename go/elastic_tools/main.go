package main

import (
	"dev/elastic_tools/common"
	"dev/elastic_tools/handler"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"path"
	"path/filepath"
)

func main() {
	config := common.GetConf()
	logPath := config.Log.Path
	if logPath == ""{
		logPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	logFile := path.Join(logPath,config.Log.Filename)
	logger := common.InitLogger(logFile,config.Log.Level)
	client,err := handler.ESClient(config.EsUrl,logFile)
	if err != nil{
		fmt.Printf("%v",err)
		os.Exit(-1)
	}
	app := cli.NewApp()
	app.Name = "operate elastic indices"
	app.Commands = []cli.Command{
		{
			Name: "delete",
			Usage: "delete indices",
			Action: func(c *cli.Context) {
				handler.DeleteIndex(client,config,logger)
			},
		},
		{
			Name: "set-tag",
			Usage: "set indices tag",
			Action: func(c *cli.Context) {
				handler.SetTag(client,config,logger)
			},
		},
		{
			Name: "repository",
			Usage: "create snapshot repository",
			Action: func(c *cli.Context) {
				repository := c.String("repository")
				handler.SnapShotRepository(client,config,repository,logger)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "repository,r",
					Usage: "snapshot repository",
					Value: "",
				},
			},
		},
		{
			Name: "snapshot",
			Usage: "create indices snapshot",
			Action: func(c *cli.Context) {
				snapshot := c.String("snapshot")
				repository := c.String("repository")
				handler.SnapShot(client,config,repository,snapshot,logger)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "snapshot,s",
					Usage: "snapshot name",
					Value: "",
				},
				cli.StringFlag{
					Name:  "repository,r",
					Usage: "where snapshot store",
					Value: "",
				},
			},
		},
		{
			Name: "get-snapshot",
			Usage: "get indices snapshot",
			Action: func(c *cli.Context) {
				repository := c.String("repository")
				handler.GetSnapShot(client,repository,logger)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "repository,r",
					Usage: "where snapshot store",
					Value: "",
				},
			},
		},
		{
			Name: "del-snapshot",
			Usage: "delete snapshot",
			Action: func(c *cli.Context) {
				snapshot := c.String("snapshot")
				repository := c.String("repository")
				handler.DeleteSnapShot(client,repository,snapshot,logger)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "snapshot,s",
					Usage: "snapshot name",
					Value: "",
				},
				cli.StringFlag{
					Name:  "repository,r",
					Usage: "where snapshot store",
					Value: "",
				},
			},
		},
		{
			Name: "get-tag",
			Usage: "get indices tag",
			Action: func(c *cli.Context) {
				indices := c.String("indices")
				handler.GetTag(client,indices,logger)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "indices,i",
					Usage: "indices name",
					Value: "",
				},
			},
		},
	}
	app.Run(os.Args)

	//repository := "my_backup"
	//snapshot := "elastic-test"
	//mapping := `{
	//	"indices": "elastic-test-2019.05.23",
	//	"ignore_unavailable": true,
	//	"include_global_state": false
	//}`
	//mapping := `{
	//       "type": "fs",
	//       "settings": {
	//           "location": "/opt/es_backup/20190523",
	//           "max_snapshot_bytes_per_sec": "50mb",
	//           "max_restore_bytes_per_sec": "100mb"
	//       }
	//   }`
	//handler.SnapShotRepository(client,repository,mapping)
	//handler.SnapShot(client,repository,snapshot,mapping)
	//handler.GetSnapShotIndex(client,repository)
	//mapping := `{
    //	"index" : {
    //    	"refresh_interval" : "2"
    //	}
	//}`
	//handler.SetTag(client,"elastic-test-2019.05.23",mapping,logger)
}
