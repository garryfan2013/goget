package rpc

import (
	_ "fmt"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/garryfan2013/goget/proxy"
	pb "github.com/garryfan2013/goget/rpc/api"
)

type RpcServer struct {
	pm proxy.ProxyManager
}

func NewRpcServer(pm proxy.ProxyManager) (*RpcServer, error) {
	return &RpcServer{pm}, nil
}

func (s *RpcServer) Serve() error {
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGoGetServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func (s *RpcServer) Add(ctx context.Context, job *pb.Job) (*pb.JobInfo, error) {
	info, err := s.pm.Add(job.Url, job.Path, job.Username, job.Passwd, int(job.Cnt))
	if err != nil {
		return nil, err
	}

	return &pb.JobInfo{
		Id:   info.Id,
		Url:  info.Url,
		Path: info.Path,
	}, nil
}

func (s *RpcServer) Get(ctx context.Context, id *pb.Id) (*pb.JobInfo, error) {
	info, err := s.pm.Get(id.Uuid)
	if err != nil {
		return nil, err
	}

	return &pb.JobInfo{
		Id:   info.Id,
		Url:  info.Url,
		Path: info.Path,
	}, nil
}

func (s *RpcServer) GetAll(e *pb.Empty, stream pb.GoGet_GetAllServer) error {
	infos, err := s.pm.GetAll()
	if err != nil {
		return err
	}

	for _, info := range infos {
		pbInfo := &pb.JobInfo{
			Id:   info.Id,
			Url:  info.Url,
			Path: info.Path,
		}
		if err := stream.Send(pbInfo); err != nil {
			return err
		}
	}

	return nil
}

func (s *RpcServer) Progress(ctx context.Context, id *pb.Id) (*pb.Stats, error) {
	stats, err := s.pm.Progress(id.Uuid)
	if err != nil {
		return nil, err
	}

	return &pb.Stats{
		Size: stats.Size,
		Done: stats.Done,
	}, nil
}

func (s *RpcServer) Start(ctx context.Context, id *pb.Id) (*pb.Empty, error) {
	err := s.pm.Start(id.Uuid)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (s *RpcServer) Stop(ctx context.Context, id *pb.Id) (*pb.Empty, error) {
	err := s.pm.Stop(id.Uuid)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (s *RpcServer) Delete(ctx context.Context, id *pb.Id) (*pb.Empty, error) {
	err := s.pm.Delete(id.Uuid)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}
