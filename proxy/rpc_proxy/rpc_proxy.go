package rpc

import (
	_ "errors"
	_ "fmt"
	"io"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/garryfan2013/goget/proxy"
	pb "github.com/garryfan2013/goget/rpc/api"
)

const (
	DefaultTimeoutSec = 30
)

func init() {
	proxy.Register(&RpcManagerBuilder{})
}

type RpcManagerBuilder struct{}

func (rmb *RpcManagerBuilder) Build() (proxy.ProxyManager, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	return &RpcManager{
		addr:  "127.0.0.1:8080",
		dopts: opts,
	}, nil
}

func (rmb *RpcManagerBuilder) Name() string {
	return proxy.ProxyRPC
}

func dialRPCServer(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

type RpcManager struct {
	addr  string
	dopts []grpc.DialOption
}

func (rm *RpcManager) Add(url string, path string, username string, passwd string, cnt int) (*proxy.JobInfo, error) {
	conn, err := dialRPCServer(rm.addr, rm.dopts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeoutSec*time.Second)
	info, err := pb.NewGoGetClient(conn).Add(ctx, &pb.Job{
		Url:      url,
		Path:     path,
		Username: username,
		Passwd:   passwd,
		Cnt:      int64(cnt),
	})
	if err != nil {
		return nil, err
	}

	return &proxy.JobInfo{
		Id:   info.Id,
		Url:  info.Url,
		Path: info.Path,
	}, nil
}

func (rm *RpcManager) Get(id string) (*proxy.JobInfo, error) {
	conn, err := dialRPCServer(rm.addr, rm.dopts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeoutSec*time.Second)
	info, err := pb.NewGoGetClient(conn).Get(ctx, &pb.Id{Uuid: id})
	if err != nil {
		return nil, err
	}

	return &proxy.JobInfo{
		Id:   info.Id,
		Url:  info.Url,
		Path: info.Path,
	}, nil
}

func (rm *RpcManager) GetAll() ([]*proxy.JobInfo, error) {
	conn, err := dialRPCServer(rm.addr, rm.dopts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeoutSec*time.Second)
	stream, err := pb.NewGoGetClient(conn).GetAll(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	var infos []*proxy.JobInfo
	for {
		info, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		infos = append(infos, &proxy.JobInfo{
			Id:   info.Id,
			Url:  info.Url,
			Path: info.Path,
		})
	}

	return infos, nil
}

func (rm *RpcManager) Progress(id string) (*proxy.Stats, error) {
	conn, err := dialRPCServer(rm.addr, rm.dopts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeoutSec*time.Second)
	stat, err := pb.NewGoGetClient(conn).Progress(ctx, &pb.Id{Uuid: id})
	if err != nil {
		return nil, err
	}

	return &proxy.Stats{
		Size: stat.Size,
		Done: stat.Done,
	}, nil
}

func (rm *RpcManager) Stop(id string) error {
	conn, err := dialRPCServer(rm.addr, rm.dopts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeoutSec*time.Second)
	_, err = pb.NewGoGetClient(conn).Stop(ctx, &pb.Id{Uuid: id})
	if err != nil {
		return err
	}

	return nil
}

func (rm *RpcManager) Delete(id string) error {
	conn, err := dialRPCServer(rm.addr, rm.dopts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, _ := context.WithTimeout(context.Background(), DefaultTimeoutSec*time.Second)
	_, err = pb.NewGoGetClient(conn).Delete(ctx, &pb.Id{Uuid: id})
	if err != nil {
		return err
	}

	return nil
}
