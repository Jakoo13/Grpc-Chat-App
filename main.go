package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/jakoo13/grpc_chat_app/pb"

	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream pb.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Server struct {
	pb.UnimplementedBroadcastServer
	Connections []*Connection
}

func (s *Server) CreateStream(pconn *pb.Connect, stream pb.Broadcast_CreateStreamServer) error {
	if s == nil {
		return fmt.Errorf("Server is nil")
	}
	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connections = append(s.Connections, conn)

	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connections {
		if conn.id == msg.ToId {
			wait.Add(1)

			go func(msg *pb.Message, conn *Connection) {
				defer wait.Done()

				if conn.active {
					err := conn.stream.Send(msg)
					fmt.Printf("Sending message to: %v from %v", conn.id, msg.ToId)

					if err != nil {
						grpcLog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
						conn.active = false
						conn.error <- err
					}
				}
			}(msg, conn)
		}

	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &pb.Close{}, nil
}

func main() {
	server := &Server{
		Connections: make([]*Connection, 0),
	}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	grpcLog.Info("Starting server at port :8080")

	pb.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
}
