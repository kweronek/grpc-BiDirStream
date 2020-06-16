package main

import (
	"context"
	"io"
	"log"
	"math/rand"

	pb "github.com/pahanini/go-grpc-bidirectional-streaming-example/src/proto"

	"time"

	"google.golang.org/grpc"
)

func main() {
	// initialize random function with setting a seed
	rand.Seed(time.Now().Unix())

	// dail server
	conn, err := grpc.Dial(":50005", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create client using pb.New<ServiceName>Client(conn)
	client := pb.NewMathClient(conn)
	stream, err := client.Max(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	var max int32
	ctx := stream.Context()
	done := make(chan bool)

	// first goroutine sends random increasing numbers to stream
	// and closes it after 10xx iterations
	go func() {
		for i := 1; i <= 100000; i++ {
			// generate random nummber and send it to stream
			rnd := int32(rand.Intn(i))
			req := pb.Request{Num: rnd}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("can not send %v", err)
			}
			log.Printf("%d sent", req.Num)
			// optional delay:			
			time.Sleep(time.Millisecond * 20)
		}
		// end transmission
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	// second goroutine receives data from stream
	// and saves result in max variable
	//
	go func() {
		for {
			resp, err := stream.Recv()
			// if stream receives EOF it closes done channel
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			max = resp.Result
			log.Printf("new max %d received", max)
		}
	}()

	// third goroutine closes done channel if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
	log.Printf("finished with max=%d", max)
}
