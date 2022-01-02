/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"log"
	"net"
	"hash/fnv"
	"fmt"
	"io"

	"google.golang.org/grpc"
	pb "helloworld/helloworld"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedChatServer
}

func (s server) Talk(srv pb.Chat_TalkServer) error {
	fmt.Println("Started server")
	ctx := srv.Context()

	for true {
		// Done returns close signal when work done on behalf of this context is complete.
		// https://stackoverflow.com/questions/3398490/checking-if-a-channel-has-a-ready-to-read-value-using-go
		select {
			case <- ctx.Done():
				// https://pkg.go.dev/context
				fmt.Println("stream closed")
				return ctx.Err()
			default:
				fmt.Println("Active")
		}

		req, err := srv.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("Error receiving message")
			continue
		}

		h := fnv.New32a()
        h.Write([]byte(req.Message))
		hString := fmt.Sprintf("hash of %s is %d", req.Message, h)

		resp := pb.ServerResponse{Message: hString}
		if err := srv.Send(&resp); err != nil {
			log.Printf("Send error %v", err)
		}
		log.Printf("Send response %v", hString)
		
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("listening...")
	s := grpc.NewServer()
	pb.RegisterChatServer(s, &server{})

	log.Printf("HOOHAH: server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
