package main

import (
	"flag"
	"fmt"
	"path"

	"github.com/garryfan2013/goget/client"
	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/manage"
	"github.com/garryfan2013/goget/record"
)

/*
	Example usage:
	goget -c 5 -o ~/Downloads https://mirrors.tuna.tsinghua.edu.cn/apache/zookeeper/stable/zookeeper-3.4.12.tar.gz

	This will start 5 go routine concurrently, each will deal with the 1/5 of the total file size,
	the successfully downloaded file will be stored at !/Downloads/zookeeper-3.4.12.tar.gz
*/

const (
	Version          = "1.0.0"
	DefaultTaskCount = 5
)

var (
	printHelp bool
	taskCount int
	savePath  string
)

func init() {
	flag.BoolVar(&printHelp, "h", false, "Printf help messages")
	flag.IntVar(&taskCount, "c", DefaultTaskCount, "Multi-task count for concurrent downloading")
	flag.StringVar(&savePath, "o", "./", "The specified download file saved path")
}

func usage() {
	fmt.Printf("goget version: %s\n", Version)
	fmt.Printf("Usage: goget [-h] [-o save_file_path] [-c task_count] remote_url\n\n")
	fmt.Printf("Options:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	/*
		Dont allow to have unrecognized flag
	*/
	if flag.NArg() > 1 {
		usage()
		return
	}

	var url string
	if url = flag.Arg(0); url == "" {
		usage()
		return
	}

	/*
		print help messages
	*/
	if printHelp == true {
		usage()
		return
	}

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

	fileName := path.Base(url)
	savePath := fmt.Sprintf("%s/%s", savePath, fileName)

	controller.SetConfig(config.KeyRemoteUrl, url)
	controller.SetConfig(config.KeyLocalPath, savePath)
	controller.SetConfig(config.KeyTaskCount, fmt.Sprintf("%d", taskCount))

	if err = controller.Start(); err != nil {
		fmt.Println(err)
		return
	}
}
