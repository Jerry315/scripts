package main

import (
	"dev/chain_check/common"
	"dev/chain_check/request"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"path"
	"path/filepath"
)

func main() {
	var logFile string
	config := common.GetConf()
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if config.Log.Path == "" {
		logFile = path.Join(basePath, config.Log.FileName)
	} else {
		logFile = path.Join(config.Log.Path, config.Log.FileName)
	}
	Logger := common.InitLogger(logFile, config.Log.Level)
	token, err := request.GetToken(config, Logger)
	if err != nil {
		Logger.Error("get token failed,exit script ")
		fmt.Println(0)
		os.Exit(0)
	}
	img4 := path.Join(basePath, "img", "mawar.jpg")
	img3 := path.Join(basePath, "img", "upload3.jpg")
	img2 := path.Join(basePath, "img", "small_02.jpg")
	img1 := path.Join(basePath, "img", "small_01.jpg")
	//request.Check2(config,token,img3,Logger)

	app := cli.NewApp()
	app.Name = "check img upload and download chain is health."
	app.Commands = []cli.Command{
		{
			Name: "event",
			Action: func(c *cli.Context) {
				result := request.Check3(config,token,img3,Logger)
				fmt.Println(result)
			},
		},
		{
			Name: "mawar",
			Action: func(c *cli.Context) {
				result := request.MawarCheck(config,token,img4,Logger)
				fmt.Println(result)
			},
		},
		{
			Name: "oss",
			Action: func(c *cli.Context) {
				result := request.Check2(config,token,img2,Logger)
				fmt.Println(result)
			},
		},
		{
			Name: "oss_key",
			Action: func(c *cli.Context) {
				result := request.Check2Key(config,token,img1,Logger)
				fmt.Println(result)
			},
		},
	}
	app.Run(os.Args)

}
