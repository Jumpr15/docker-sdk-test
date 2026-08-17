package main

import (
	"context"
	"log"
	"net/netip"
	// "fmt"
	// "io"
	// "os"

	// "github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

func main() {
	ctx := context.Background()
	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	defer apiClient.Close()

	imageName := "envoy-test-app:latest"

	envVars := []string{"PORT=80"}

	port, ok := network.PortFrom(80, "tcp")
	if !ok {
		log.Fatal("Problem creating port")
	}

	exposedPorts := network.PortSet{
		port: struct{}{},
	}

	hostPort, err := netip.ParseAddr("127.0.0.1")
	if err != nil {
		log.Fatal(err)
	}

	portBinding := network.PortBinding{
		HostIP: hostPort,
		HostPort: "80",
	}

	portBindings := network.PortMap{
		port: []network.PortBinding{portBinding},
	}

	res, err := apiClient.ContainerCreate(ctx, client.ContainerCreateOptions{
		Name: "Docker-webserver-container",
		Image: imageName,
		Config: &container.Config{
			Env: envVars,
			ExposedPorts: exposedPorts,
		},
		HostConfig: &container.HostConfig{
			PortBindings: portBindings,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Container ID is:", res.ID)

	_, err = apiClient.ContainerStart(ctx, res.ID, client.ContainerStartOptions{})
	if err != nil {
		log.Fatal(err)
	}

}