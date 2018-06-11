package main

import (
	"flag"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/garryfan2013/goget/manager"
	pb "github.com/garryfan2013/goget/rpc/api"
)

/*
	Example usage:
	cli -c 5 -o ./ https://mirrors.tuna.tsinghua.edu.cn/apache/zookeeper/stable/zookeeper-3.4.12.tar.gz

	This will start 5 go routine concurrently, each will deal with the 1/5 of the total file size,
	the successfully downloaded file will be stored at !/Downloads/zookeeper-3.4.12.tar.gz
*/

const (
	Version = "1.0.0"
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
	flag.IntVar(&taskCount, "c", manager.DefaultTaskCount, "Multi-task count for concurrent downloading")
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

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial("127.0.0.1:8080", opts...)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewGoGetClient(conn)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	id, err := client.Add(ctx, &pb.Job{
		Url:      urlStr,
		Path:     savePath,
		Username: userName,
		Passwd:   passwd,
		Cnt:      int64(taskCount)})
	if err != nil {
		fmt.Printf("JobManager Add Job failed: %s\n", err.Error())
		return
	}

	var count int
	for {
		<-time.After(time.Millisecond * 500)
		count += 1

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		stats, err := client.Progress(ctx, id)
		if err != nil {
			fmt.Println(err)
			return
		}

		if stats.Size > 0 {
			fmt.Printf("\rJob progress: %d/%d %dkb/s", stats.Done, stats.Size, stats.Done/int64(count*1024/2))
			if stats.Size == stats.Done {
				fmt.Printf("\nJob Done\n")
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				client.Stop(ctx, id)
				return
			}
		}
	}
}