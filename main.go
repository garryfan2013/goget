package main

import (
	"flag"
	"fmt"
	"net/url"
	"path"

	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/manager"
	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"
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

	SchemeHttp  = "http"
	SchemeHttps = "https"
	SchemeFtp   = "ftp"
)

var (
	printHelp bool
	taskCount int
	savePath  string
	userName  string
	passwd    string
)

func init() {
	flag.BoolVar(&printHelp, "h", false, "Printf help messages")
	flag.IntVar(&taskCount, "c", DefaultTaskCount, "Multi-task count for concurrent downloading")
	flag.StringVar(&savePath, "o", "./", "The specified download file saved path")
	flag.StringVar(&userName, "u", "", "username for authentication")
	flag.StringVar(&passwd, "p", "", "passwd for authentication")
}

func usage() {
	fmt.Printf("goget version: %s\n", Version)
	fmt.Printf("Usage: goget [-h] [-o save_file_path] [-c task_count] [-u username] [-p passwd] remote_url\n\n")
	fmt.Printf("Options:\n")
	flag.PrintDefaults()
}

func getProtocol(urlStr string) (int, error) {
	fmtUrl, err := url.Parse(urlStr)
	if err != nil {
		return -1, err
	}

	switch fmtUrl.Scheme {
	case SchemeHttp:
	case SchemeHttps:
		return source.HttpProtocol, nil
	case SchemeFtp:
		return source.FtpProtocol, nil
	default:
	}

	return -1, fmt.Errorf("Unsupported url scheme: %s", fmtUrl)
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

	var urlStr string
	if urlStr = flag.Arg(0); urlStr == "" {
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

	var src source.StreamReader
	var snk sink.StreamWriter
	var controller manager.Controller
	var err error
	var protocol int

	protocol, err = getProtocol(urlStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	src, err = source.NewStreamReader(protocol)
	if err != nil {
		fmt.Println(err)
		return
	}

	snk, err = sink.NewStreamWriter(sink.LocalFileType)
	if err != nil {
		fmt.Println(err)
		return
	}

	controller, err = manager.NewController(manager.MultiTaskType)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = controller.Open(src, snk); err != nil {
		fmt.Println(err)
		return
	}
	defer controller.Close()

	savePath := fmt.Sprintf("%s/%s", savePath, path.Base(urlStr))

	controller.SetConfig(config.KeyRemoteUrl, urlStr)
	controller.SetConfig(config.KeyLocalPath, savePath)
	controller.SetConfig(config.KeyTaskCount, fmt.Sprintf("%d", taskCount))
	if userName != "" {
		controller.SetConfig(config.KeyUserName, userName)
	}
	if passwd != "" {
		controller.SetConfig(config.KeyPasswd, passwd)
	}

	fmt.Println("Download Started... please wait")
	if err = controller.Start(); err != nil {
		fmt.Println(err)
		return
	}
}
