package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	cmlclient "github.com/rschmied/gocmlclient"
)

func main() {
	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource: true,
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC822,
		}),
	))

	// address and lab id
	host, found := os.LookupEnv("CML_HOST")
	if !found {
		slog.Error("CML_HOST env var not found!")
		return
	}
	// labID, found := os.LookupEnv("CML_LABID")
	// if !found {
	// 	slog.Error("CML_LABID env var not found!")
	// 	return
	// }
	// _ = labID

	// auth related
	username, user_found := os.LookupEnv("CML_USERNAME")
	password, pass_found := os.LookupEnv("CML_PASSWORD")
	token, token_found := os.LookupEnv("CML_TOKEN")
	if !(token_found || (user_found && pass_found)) {
		slog.Error("either CML_TOKEN or CML_USERNAME and CML_PASSWORD env vars must be present!")
		return
	}
	ctx := context.Background()
	client := cmlclient.New(host, false, false)
	// if err := client.Ready(ctx); err != nil {
	// 	log.Fatal(err)
	// }
	if token_found {
		client.SetToken(token)
	} else {
		client.SetUsernamePassword(username, password)
	}

	// cert, err := os.ReadFile("ca.pem")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// err = client.SetCACert(cert)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// topo, err := os.ReadFile("topology.yaml")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// l, err := client.ImportLab(ctx, string(topo))

	// l, err := client.LabGet(ctx, labID, false)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// nd, err := client.GetNodeDefs(ctx)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// for _, n := range nd {
	// 	a := n.Device.Interfaces.DefaultCount
	// 	log.Println(a)
	// }

	// nd, err := client.GetImageDefs(ctx)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// result, err := client.UserGroups(ctx, "cc42bd56-1dc6-445c-b7e7-569b0a8b0c94")
	err := client.Ready(ctx)
	if errors.Is(err, cmlclient.ErrSystemNotReady) {
		slog.Error("it is not ready")
		return
	}
	if err != nil && !errors.Is(err, cmlclient.ErrSystemNotReady) {
		slog.Error("ready", slog.Any("error", err))
		return
	}
	node := &cmlclient.Node{
		// ID:    "28ec08ec-483a-415a-a3ed-625b0d45bef0",
		// ID:    "8116a609-8b68-4e0f-a196-5225da9f05c0",
		ID:    "0577f1c4-4907-4c49-a4fd-c6daa61b6e78",
		LabID: "2b7435f2-b247-4cc8-8509-6b0d0f593c4c",
	}
	node, err = client.NodeGet(ctx, node, false)
	if err != nil {
		slog.Error("nodeget", slog.Any("error", err))
		return
	}

	je, err := json.Marshal(node)
	if err != nil {
		slog.Error("marshal", slog.Any("error", err))
		return
	}
	fmt.Println(string(je))

	// lab, err := client.LabGet(ctx, "2b7435f2-b247-4cc8-8509-6b0d0f593c4c", true)
	// if err != nil {
	// 	slog.Error("get", slog.Any("error", err))
	// 	return
	// }
	//
	// for _, v := range lab.Nodes {
	// 	if v.Configuration != nil {
	// 		fmt.Printf("[1] %T: %s\n", v.Configuration, *v.Configuration)
	// 	}
	// 	fmt.Printf("[2] %T: %+v\n", v.Configurations, v.Configurations)
	// }
	// return
	// je, err := json.Marshal(lab)
	// if err != nil {
	// 	slog.Error("marshal", slog.Any("error", err))
	// 	return
	// }
	// fmt.Println(string(je))
}
