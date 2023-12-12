package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"

	svc "arbitrum-publisher/internal/service"
	"arbitrum-publisher/pkg/ipc"
	svcnats "arbitrum-publisher/pkg/nats"
)

func main() {
	flagIpcPath := flag.String("socket", "", "Arbitrum node URI to establish IPC/WebSocket connection")
	flagNatsUrls := flag.String("nats", "", "NATS server URLs (separated by comma)")
	flagUserCredsSeed := flag.String("nats-nkey", "", "NATS NKey string")
	flagPrefixOrg := flag.String("stream-prefix", "", "Streams prefix")
	flagNodeNetwork := flag.String("stream-network-infix", "", "Arbitrum network stream infix, e.g.: mainnet, ")
	flag.Parse()

	if *flagIpcPath == "" {
		log.Fatal("Socket URI is required!")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	opts := []nats.Option{}

	flagUserCredsJWT, err := svcnats.CreateAppJwt(*flagUserCredsSeed)
	if err != nil {
		log.Fatalf("failed to create sub JWT: %v", err)
	}

	opts = append(opts, nats.UserJWTAndSeed(flagUserCredsJWT, *flagUserCredsSeed))

	svcn := svcnats.MustConnect(
		svcnats.Config{
			URI:  *flagNatsUrls,
			Opts: opts,
		})
	log.Println("NATS server connected.")

	service, errSvc := ipc.NewClient(ctx, *flagIpcPath)
	if errSvc != nil {
		log.Fatalf("Starting service error: %s", errSvc.Error())
	}
	log.Println("IPC socket connected.")

	pubService := svc.New(ctx, service, svcn, *flagPrefixOrg, *flagNodeNetwork)
	errCh := pubService.Run()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		log.Fatal(err)
	}
}
