package main

import (
	"fmt"

	"github.com/garryfan2013/goget/client"
	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/manage"
	"github.com/garryfan2013/goget/record"
)

const (
	FileUrl         = "https://mirrors.tuna.tsinghua.edu.cn/apache/zookeeper/stable/zookeeper-3.4.12.tar.gz"
	FilePath        = "./zookeeper-3.4.12.tar.gz"
	ConcurrentCount = "5"
)

func main() {
	var url, path string = FileUrl, FilePath
	var taskCount string = ConcurrentCount

	source, err := client.NewCrawler(client.HttpProtocol)
	if err != nil {
		fmt.Println(err)
		return
	}

	sink, err := record.NewHandler(record.LocalFileType)
	if err != nil {
		fmt.Println(err)
		return
	}

	controller, err := manage.NewController(manage.MultiTaskType)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer controller.Close()

	if err = controller.Open(source, sink); err != nil {
		fmt.Println(err)
		return
	}

	controller.SetConfig(config.KeyRemoteUrl, url)
	controller.SetConfig(config.KeyLocalPath, path)
	controller.SetConfig(config.KeyTaskCount, taskCount)

	if err = controller.Start(); err != nil {
		fmt.Println(err)
		return
	}
}
