package main

import (
	"fmt"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/garryfan2013/goget/manager"
	pb "github.com/garryfan2013/goget/rpc/api"
)

type GogetServer struct {
	jobManager *manager.JobManager
}

func (s *GogetServer) Add(ctx context.Context, job *pb.Job) (*pb.Id, error) {
	id, err := s.jobManager.Add(job.Url, job.Path, job.Username, job.Passwd, int(job.Cnt))
	if err != nil {
		return nil, err
	}

	return &pb.Id{Uuid: id.String()}, nil
}

func (s *GogetServer) Progress(ctx context.Context, id *pb.Id) (*pb.Stats, error) {
	jid, err := manager.FromString(id.Uuid)
	if err != nil {
		return nil, err
	}

	stats, err := s.jobManager.Progress(jid)
	if err != nil {
		return nil, err
	}

	return &pb.Stats{Size: stats.Size, Done: stats.Done}, nil
}

func (s *GogetServer) Stop(ctx context.Context, id *pb.Id) (*pb.Stats, error) {
	jid, err := manager.FromString(id.Uuid)
	if err != nil {
		return nil, err
	}

	err = s.jobManager.Stop(jid)
	if err != nil {
		return nil, err
	}

	return &pb.Stats{}, nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	gogetServer := &GogetServer{jobManager: manager.GetInstance()}
	grpcServer := grpc.NewServer()
	pb.RegisterGoGetServer(grpcServer, gogetServer)
	grpcServer.Serve(lis)
}
