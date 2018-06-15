package main

import (
	"flag"
	"fmt"

	"github.com/garryfan2013/goget/proxy"
	_ "github.com/garryfan2013/goget/proxy/local"
	_ "github.com/garryfan2013/goget/proxy/rpc_proxy"
)

/*
	Example usage:
	./cli.exe -c 5 -o ./ https://mirrors.tuna.tsinghua.edu.cn/apache/zookeeper/stable/zookeeper-3.4.12.tar.gz

	This will start 5 go routine concurrently, each will deal with the 1/5 of the total file size,
	the successfully downloaded file will be stored at !/Downloads/zookeeper-3.4.12.tar.gz
*/

const (
	Version = "1.0.0"
)

var (
	listAllJobs bool

	listSingle    bool
	queryProgress bool
	deleteSingle  bool
	id            string

	printHelp bool

	taskCount int
	savePath  string
	userName  string
	passwd    string
)

func init() {
	flag.BoolVar(&listAllJobs, "L", false, "List all current jobs' information")

	flag.BoolVar(&listSingle, "l", false, "List single job's information")
	flag.BoolVar(&deleteSingle, "d", false, "Delete single job")
	flag.StringVar(&id, "i", "", "job id")

	flag.BoolVar(&queryProgress, "q", false, "Query progress of job by id")

	flag.BoolVar(&printHelp, "h", false, "Printf help messages")

	flag.IntVar(&taskCount, "c", 5, "Multi-task count for concurrent downloading")
	flag.StringVar(&savePath, "o", "./", "The specified download file saved path")
	flag.StringVar(&userName, "u", "", "username for authentication")
	flag.StringVar(&passwd, "p", "", "passwd for authentication")
}

func usage() {
	fmt.Printf("goget version: %s\n", Version)
	fmt.Printf("Usage: goget [-L] [-l -i id] [-h] [[-o save_file_path] [-c task_count] [-u username] [-p passwd] remote_url]\n\n")
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

	/*
		print help messages
	*/
	if printHelp == true {
		usage()
		return
	}

	pm, err := proxy.GetProxyManager(proxy.ProxyRPC)
	if err != nil {
		fmt.Println(err)
		return
	}
	/*
		List all jobs
	*/
	if listAllJobs == true {
		jobs, err := pm.GetAll()
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, job := range jobs {
			fmt.Printf("Id: %s, Url: %s, Path: %s\n", job.Id, job.Url, job.Path)
		}
		return
	}

	/*
		List single job with id
	*/
	if listSingle == true {
		if id == "" {
			usage()
			return
		}

		job, err := pm.Get(id)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Id: %s, Url: %s, Path: %s\n", job.Id, job.Url, job.Path)
		return
	}

	/*
		Delete single job with id
	*/
	if deleteSingle == true {
		if id == "" {
			usage()
			return
		}

		err := pm.Delete(id)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Job-%s deleted\n", id)
		return
	}

	/*
		Get the progress info for job with id
	*/
	if queryProgress == true {
		if id == "" {
			usage()
			return
		}

		stat, err := pm.Progress(id)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Job-%s progress: %d/%d\n", id, stat.Done, stat.Size)
		return
	}

	/*
		Here we deal with the download job
	*/
	var urlStr string
	if urlStr = flag.Arg(0); urlStr == "" {
		usage()
		return
	}

	_, err = pm.Add(urlStr, savePath, userName, passwd, taskCount)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
