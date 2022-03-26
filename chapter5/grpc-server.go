package main

import "context"

//定义服务结构体
type ProgrammerServiceServer struct{}

func (p *ProgrammerServiceServer) GetProgrammerInfo(ctx context.Context, req *pb.Request) (resp *pb.R)
