package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

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
	listAllJobs bool

	listSingle    bool
	queryProgress bool
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

func dialRPCServer() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial("127.0.0.1:8080", opts...)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return conn, nil
}

func listAll() {
	conn, err := dialRPCServer()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewGoGetClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	stream, err := client.GetAll(ctx, &pb.Empty{})
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		info, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(info)
	}
}

func list(jid string) {
	conn, err := dialRPCServer()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewGoGetClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	info, err := client.Get(ctx, &pb.Id{Uuid: jid})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(info)
}

func progress(jid string) {
	conn, err := dialRPCServer()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewGoGetClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	stat, err := client.Progress(ctx, &pb.Id{Uuid: jid})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(stat)
}

func Add(url string, path string, username string, passwd string, cnt int) {
	conn, err := dialRPCServer()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewGoGetClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	info, err := client.Add(ctx, &pb.Job{
		Url:      url,
		Path:     path,
		Username: username,
		Passwd:   passwd,
		Cnt:      int64(cnt)})
	if err != nil {
		fmt.Printf("JobManager Add Job failed: %s\n", err.Error())
		return
	}

	fmt.Println(info)
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

	/*
		List all jobs
	*/
	if listAllJobs == true {
		listAll()
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

		list(id)
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

		progress(id)
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

	Add(urlStr, savePath, userName, passwd, taskCount)
	return
}
