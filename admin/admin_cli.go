package main

import (
    "os"
    "fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
    "flag"
    "log"
    "context"
    "time"
    pb "irio/linter_proto"
)

var (
    listen_addr = flag.String("address", "0.0.0.0:1337", "The Machine Spawner address")
    addVersion = flag.NewFlagSet("add_version", flag.ExitOnError)
    containerUrl = addVersion.String("container_url", "(invalid)", "The container url with linter")
    language = addVersion.String("language", "", "Language to be used")
    major = addVersion.Int("major", -1, "Major")
    minor = addVersion.Int("minor", -1, "Minor")
    patch = addVersion.Int("patch", -1, "Patch")
)

func add_version(args []string) {
    addVersion.Parse(args)
    if (*major < 0 || *minor < 0 || *patch < 0) {
        log.Fatalf("Version format illegal: (%d:%d:%d)", *major, *minor, *patch)
    }

    linterVersion := pb.LinterVersion {
        Language: *language,
        Major: int32(*major),
        Minor: int32(*minor),
        Patch: int32(*patch),
    }

    transportCreds := grpc.WithTransportCredentials(insecure.NewCredentials())

    conn, err := grpc.Dial(*listen_addr, transportCreds)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMachineSpawnerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

    r, err := c.AddLinter(ctx, &pb.AddLinterRequest{ContainerUrl: *containerUrl, LinterVersion: &linterVersion})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.String())
}

func main() {
    if len(os.Args) == 1 {
        // TODO: binarki nazwę mi wstaw
        fmt.Println("usage: admin <command> [<args>]")
        fmt.Println("subcommands: ")
        fmt.Println(" add_version           Adds new version")
        return
    }
    *listen_addr = os.Args[1]

    switch os.Args[2] {
    case "add_version":
        add_version(os.Args[3:])
    default:
        fmt.Printf("%q is not valid command.\n", os.Args[1])
        os.Exit(2)
    }
}
