package main

import (
	"dev/backup_oss/common"
	"dev/backup_oss/request"
	"fmt"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func upload(conf common.Yaml, token string,logger *zap.Logger) {
	uploadPath := conf.Upload.Path
	if uploadPath == "" {
		uploadPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	ts := time.Now().Format(conf.Upload.Layout)
	var upFiles []string
	files, _ := ioutil.ReadDir(uploadPath)
	for _, f := range files {
		if strings.Contains(f.Name(),ts) {
			upFiles = append(upFiles, path.Join(uploadPath, f.Name()))
		}
	}
	for _, upFile := range upFiles {
		if upFile != "" {
			err := request.UploadFile(conf, token, "file", upFile,logger)
			if err != nil{
				fmt.Printf("%#v", err)
				logger.Error(fmt.Sprintf("upload file: %s failed",upFile))
			}
		}else {
			fmt.Println("have no files to upload")
			logger.Warn("have no files to upload")
		}
	}
}

func download(conf common.Yaml, token string,logger *zap.Logger) {
	objIds := conf.Download.ObjIds
	for _, objId := range objIds {
		if objId != "" {
			request.DownloadFile(conf, token, objId,logger)
		}else {
			fmt.Println("have no files to download")
			logger.Warn("have no files to download")
		}
	}
}

func main() {
	conf := common.GetConf()
	logger := common.InitLogger()
	token := request.GetToken(conf,logger)
	app := cli.NewApp()
	app.Name = "backup object to oss"
	app.Commands = []cli.Command{
		{
			Name: "upload",
			Action: func(c *cli.Context) {
				upload(conf, token,logger)
			},
		},
		{
			Name: "download",
			Action: func(c *cli.Context) {
				download(conf, token,logger)
			},
		},
	}
	app.Run(os.Args)
}
