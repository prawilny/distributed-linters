package main

import (
    "context"
    "log"
    "time"
    "google.golang.org/grpc"
    pb "irio/linter_proto"
)

func main() {
    var opts []grpc.DialOption
    opts = append(opts, grpc.WithBlock())
    opts = append(opts, grpc.WithInsecure())
    peerAddr := "localhost:22222"
    log.Printf("Dialing: %v", peerAddr)
    conn, err := grpc.Dial(peerAddr, opts...)
    log.Printf("Connected to: %v", peerAddr)
    if err != nil {
        log.Fatalf("fail to dial: %v", err)
    }
    defer conn.Close()
    client := pb.NewLoadBalancerClient(conn)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err = client.SetConfig(ctx, &pb.SetConfigRequest{
        Workers: []*pb.Worker{
            &pb.Worker{
                Address: "localhost",
                Port: 11111,
                Attrs: &pb.LinterAttributes{
                    Version: "1.0",
                    Language: "python",
                },
            },
            &pb.Worker{
                Address: "localhost",
                Port: 11112,
                Attrs: &pb.LinterAttributes{
                    Version: "1.1",
                    Language: "python",
                },
            },
        },
        Weights: []*pb.Weight{
            &pb.Weight{
                Weight: 1.0,
                Attrs: &pb.LinterAttributes{
                    Version: "1.0",
                    Language: "python",
                },
            },
            &pb.Weight{
                Weight: 2.0,
                Attrs: &pb.LinterAttributes{
                    Language: "python",
                    Version: "1.1",
                },
            },
        },
    })
    if err != nil {
        log.Fatalf("%v.SetConfig(_) = _, %v: ", client, err)
    }
}
