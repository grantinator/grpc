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

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"time"
	"fmt"
	"os"
	"bufio"
	"strings"

	"google.golang.org/grpc"
	pb "helloworld/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	rand.Seed(time.Now().Unix())

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// defer conn.Close()
	client := pb.NewChatClient(conn)
	stream, err := client.Talk(context.Background())
	if err != nil {
		log.Fatalf("error opening stream %v", err)
	}

	ctx := stream.Context()
	done := make(chan bool)

	// goroutine to send messages to stream until quit message
	go func() {
		buf := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			msg, err := buf.ReadString('\n')
			msg = strings.Replace(msg, "\n", "", -1)
			if err != nil {
				fmt.Println(err)
			} else {
				req := pb.ClientMessage{Message: msg}
				err := stream.Send(&req)
				if err != nil {
					log.Fatalf("Can not send %v", err)
				}
				log.Printf("%s sent", msg)
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()

	// goroutine to listen for messages from stream
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("Can not receive %v", err)
			}
			msg := resp.Message
			log.Printf("Message from server: %s", msg)
		}
	}()

	// closes done channel if context is done
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		if err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
	log.Printf("Closed stream")

}
