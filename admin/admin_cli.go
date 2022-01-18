package main

import (
	"context"
	"flag"
	"fmt"
	pb "irio/linter_proto"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	listen_addr = flag.String("address", "0.0.0.0:1337", "machine manager address")

	addVersion   = flag.NewFlagSet("add_version", flag.ExitOnError)
	imageUrl     = addVersion.String("image_url", "(invalid)", "The linter container image url")
	language_add = addVersion.String("language", "forth", "Language to be used")
	version_add  = addVersion.String("version", "(optimized out)", "Version")

	removeVersion   = flag.NewFlagSet("remove_version", flag.ExitOnError)
	language_remove = removeVersion.String("language", "forth", "Language to be used")
	version_remove  = removeVersion.String("version", "(optimized out)", "Version")

	listVersions  = flag.NewFlagSet("list_versions", flag.ExitOnError)
	language_list = listVersions.String("language", "forth", "Language to show versions")

	transportCreds = grpc.WithTransportCredentials(insecure.NewCredentials())
)

func dial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, transportCreds)
	if err != nil {
		log.Fatalf("could not connect to %s: %v", addr, err)
	}
	return conn
}

func add_version(args []string) {
	addVersion.Parse(args)
	if *language_add == "" || *version_add == "" {
		log.Fatal("Language and version cannot be empty")
	}

	linterVersion := pb.LinterAttributes{
		Language: *language_add,
		Version:  *version_add,
	}
	linterRequest := pb.AppendLinterRequest{
		ImageUrl: *imageUrl,
		Attrs:    &linterVersion,
	}

	conn := dial(*listen_addr)
	defer conn.Close()
	c := pb.NewMachineManagerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	r, err := c.AppendLinter(ctx, &linterRequest)
	if err != nil {
		log.Fatalf("adding linter failed: %v", err)
	}
	log.Printf("Got response: %s", r.String())
}

func remove_version(args []string) {
	removeVersion.Parse(args)
	if *language_remove == "" || *version_remove == "" {
		log.Fatal("Language and version cannot be empty")
	}

	linterVersion := pb.LinterAttributes{
		Language: *language_remove,
		Version:  *version_remove,
	}
	conn := dial(*listen_addr)
	defer conn.Close()
	c := pb.NewMachineManagerClient(conn)

	ctx, ctx_cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer ctx_cancel()
	r, err := c.RemoveLinter(ctx, &linterVersion)
	if err != nil {
		log.Fatalf("removing linter failed: %v", err)
	}
	log.Printf("Got response: %s", r.String())
}

func set_proportions(args []string) {
	if len(args)%3 != 0 {
		log.Fatal("need triples (language, version, proportion)")
	}
	cmdArgs := []*pb.Weight{}
	for i := 0; i < len(args); i += 3 {
		f, err := strconv.ParseFloat(args[i+2], 32)
		if err != nil || f <= 0 {
			log.Fatalf("%s is not a real positive float", args[i+2])
		}
		cmdArgs = append(cmdArgs, &pb.Weight{
			Attrs:  &pb.LinterAttributes{Language: args[i], Version: args[i+1]},
			Weight: float32(f),
		})
	}

	conn := dial(*listen_addr)
	defer conn.Close()
	c := pb.NewMachineManagerClient(conn)

	ctx, ctx_cancel := context.WithTimeout(context.Background(), time.Second)
	defer ctx_cancel()
	r, err := c.SetProportions(ctx, &pb.LoadBalancingProportions{Weights: cmdArgs})
	if err != nil {
		log.Fatalf("setting proportions failed: %v", err)
	}
	log.Printf("Got response: %s", r.String())
}

func list_versions(args []string) {
	listVersions.Parse(args)

	conn := dial(*listen_addr)
	defer conn.Close()
	c := pb.NewMachineManagerClient(conn)

	ctx, ctx_cancel := context.WithTimeout(context.Background(), time.Second)
	defer ctx_cancel()
	r, err := c.ListVersions(ctx, &pb.Language{Language: *language_list})
	if err != nil {
		log.Fatalf("listing versions failed: %v", err)
	}
	log.Printf("Got response: %s", r.String())
}

func main() {
	if len(os.Args) < 3 {
		// TODO: binarki nazwę mi wstaw
		fmt.Println("usage: admin <machine_manager_address> <command> [<args>]")
		fmt.Println("subcommands: ")
		fmt.Println(" add_version           Adds new version")
		fmt.Println(" remove_version        Removes preexisting version")
		fmt.Println(" list_versions         Lists existing versions")
		fmt.Println(" set_proportions       Set proportions for already existing versions")
		return
	}
	*listen_addr = os.Args[1]
	rest_args := os.Args[3:]

	switch os.Args[2] {
	case "add_version":
		add_version(rest_args)
	case "remove_version":
		remove_version(rest_args)
	case "set_proportions":
		set_proportions(rest_args)
	case "list_versions":
		list_versions(rest_args)
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[2])
		os.Exit(2)
	}
}
