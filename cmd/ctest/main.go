package main

import (
	"context"
	"errors"
	"log"
	"os"

	cmlclient "github.com/rschmied/gocmlclient"
)

func main() {
	// address and lab id
	host, found := os.LookupEnv("CML_HOST")
	if !found {
		log.Println("CML_HOST env var not found!")
		return
	}
	// labID, found := os.LookupEnv("CML_LABID")
	// if !found {
	// 	log.Println("CML_LABID env var not found!")
	// 	return
	// }
	// _ = labID

	// auth related
	username, user_found := os.LookupEnv("CML_USERNAME")
	password, pass_found := os.LookupEnv("CML_PASSWORD")
	token, token_found := os.LookupEnv("CML_TOKEN")
	if !(token_found || (user_found && pass_found)) {
		log.Println("either CML_TOKEN or CML_USERNAME and CML_PASSWORD env vars must be present!")
		return
	}
	ctx := context.Background()
	client := cmlclient.New(host, true, false)
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
		log.Println("it is not ready")
	}
	if err != nil && !errors.Is(err, cmlclient.ErrSystemNotReady) {
		log.Println(err)
		return
	}

	// je, err := json.Marshal(result)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(string(je))
}
