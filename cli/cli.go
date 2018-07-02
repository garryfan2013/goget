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
	./cli.exe -a https://mirrors.tuna.tsinghua.edu.cn/apache/zookeeper/stable/zookeeper-3.4.12.tar.gz

	This will start 5 go routine concurrently, each will deal with the 1/5 of the total file size,
	the successfully downloaded file will be stored at !/Downloads/zookeeper-3.4.12.tar.gz
*/

const (
	Version = "1.0.0"
)

var (
	listAllJobs bool

	startJob      bool
	stopJob       bool
	listSingle    bool
	queryProgress bool
	deleteSingle  bool

	printHelp bool

	jobUrl    string
	taskCount int
	savePath  string
	userName  string
	passwd    string
)

func init() {
	flag.BoolVar(&listAllJobs, "L", false, "List all current jobs' information")

	flag.BoolVar(&listSingle, "l", false, "List single job's information")
	flag.BoolVar(&deleteSingle, "d", false, "Delete a specified job")
	flag.BoolVar(&startJob, "s", false, "Start a specified job")
	flag.BoolVar(&stopJob, "S", false, "Stop a specified job")
	flag.BoolVar(&queryProgress, "q", false, "Query progress of job by id")

	flag.BoolVar(&printHelp, "h", false, "Printf help messages")

	flag.StringVar(&jobUrl, "a", "", "To add a new job")
	flag.IntVar(&taskCount, "c", 5, "Multi-task count for concurrent downloading")
	flag.StringVar(&savePath, "o", "./", "The specified download file saved path")
	flag.StringVar(&userName, "u", "", "username for authentication")
	flag.StringVar(&passwd, "p", "", "passwd for authentication")
}

func usage() {
	fmt.Printf("goget version: %s\n", Version)
	fmt.Printf("Usage: goget [-L] [-sSlq id] [-h] [-a remote_url [-o save_file_path] [-c task_count] [-u username] [-p passwd] ]\n\n")
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
	if printHelp {
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
	if listAllJobs {
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
	if listSingle {
		job, err := pm.Get(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Id: %s, Url: %s, Path: %s\n", job.Id, job.Url, job.Path)
		return
	}

	/*
		Start single job with id
	*/
	if startJob {
		err := pm.Start(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Job-%s started\n", flag.Arg(0))
		return
	}

	/*
		Stop single job with id
	*/
	if stopJob {
		err := pm.Stop(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Job-%s stopped\n", flag.Arg(0))
		return
	}

	/*
		Delete single job with id
	*/
	if deleteSingle == true {
		err := pm.Delete(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Job-%s deleted\n", flag.Arg(0))
		return
	}

	/*
		Get the progress info for job with id
	*/
	if queryProgress == true {
		stat, err := pm.Progress(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Job-%s progress: %d %d %d/%d\n", flag.Arg(0), stat.Status, stat.Rate, stat.Done, stat.Size)
		return
	}

	/*
		Here we deal with the new job
	*/
	if jobUrl == "" {
		usage()
		return
	}

	info, err := pm.Add(jobUrl, savePath, userName, passwd, taskCount)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("job-%s added\n", info.Id)
}
