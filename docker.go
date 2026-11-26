package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type ContainerLifecycle interface {
	CreateContainer(ContainerConfig)
	StopContainer()
}

type ContainerConfig struct {
	ImageName string
	ContainerName string
	EnvMap map[string]any
	PortConfig PortConfig
}

func NewContainerConfig(imageName string, containerName string, EnvMap map[string]any, PortConfig PortConfig) (ContainerConfig) {
	config := ContainerConfig{
		ImageName: imageName,
		ContainerName: containerName,
		EnvMap: EnvMap,
		PortConfig: PortConfig,
	}
	return config
}

type Containerlet struct {
	ContainerConfig
	ContainerLifecycle
}

// creates container and if successful returns channel to listen to container events a.k.a detect failures
func (clet *Containerlet) CreateAndListen() {}

func NewContainerlet(ContainerConfig ContainerConfig, ContainerLifecycle ContainerLifecycle) (*Containerlet) {
	client := Containerlet{
		ContainerConfig: ContainerConfig,
		ContainerLifecycle: ContainerLifecycle,
	}
	return &client
}

type MobyClient struct {
	Ctx context.Context 
	Client *client.Client
}

func (mb *MobyClient) CreateContainer(cfg ContainerConfig) {
	envVars :=  MapToEnvArr(cfg.EnvMap)
	
	exposedPorts, portBindings := ConvertPortConfig([]PortConfig{cfg.PortConfig}) // not good practices change later !!! for port flexibility


	res, err := mb.Client.ContainerCreate(mb.Ctx, client.ContainerCreateOptions{
		Name: cfg.ContainerName,
		Image: cfg.ImageName,
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

	log.Println(res) // res.ID is string for querying for container metadata
}

func (mb *MobyClient) StopContainer() {}

func NewMobyClient(ctx context.Context) (*MobyClient, error) {
	c, err := client.New() // use fromenv later to inject env variables
	// defer client.Close()
	mobyClient := MobyClient{
		Ctx: ctx,
		Client: c, // dont forget to close 
	}
	return &mobyClient, err
}

type PortConfig struct {
	Protocol network.IPProtocol // represented by string type
	Address netip.Addr
	HostPort network.Port
	ExposedPort network.Port
}

func NewPortConfig(Address string, HostPort uint16, ExposedPort uint16, Protocol network.IPProtocol) (PortConfig) {
	address, err := netip.ParseAddr(Address)
	if err != nil {
		log.Fatal(err)
	}
	hostPort, ok := network.PortFrom(HostPort, Protocol)
	if ok == false  {
		log.Fatal("Error occurred during converting HostInt to HostPort")
	}
	exposedPort, ok := network.PortFrom(ExposedPort, Protocol)	
	if ok == false {
		log.Fatal("Error occurred during converting ExposedInt to ExposedPort")
	}
	portConfig := PortConfig{
		Protocol: Protocol,
		Address: address,
		HostPort: hostPort,
		ExposedPort: exposedPort,
	}
	return portConfig
}

// converts map of ENV_VAR_NAME: ANY_TYPE_VALUE to passable array into docker env config as arg
func MapToEnvArr(envMap map[string]any) []string {
	var envArr []string
	var envVar string
	for k, v := range envMap {
		envVar = fmt.Sprintf("%s=%s", k, fmt.Sprint(v))
		envArr = append(envArr, envVar)
	}
	return envArr
}

// converts string list of ports to expose for a container to docker Config/HostConfig compatible types to pass as args
func ConvertPortConfig(portConfigList []PortConfig) (network.PortSet, network.PortMap) { 
	fmt.Println(portConfigList)

	exposedPorts := make(map[network.Port]struct{}) // map of type network.PortSet
	portBindings := make(map[network.Port][]network.PortBinding) // map of type network.PortMap

	for _, portConfig := range portConfigList {

		exposedPorts[portConfig.ExposedPort] = struct{}{} 
		portBindings[portConfig.HostPort] = []network.PortBinding{{
			HostIP: portConfig.Address,
			HostPort: portConfig.HostPort.String(),
		}}
	}

	return exposedPorts, portBindings

}

func main() {
	imageName := "envoy-test-app:latest"
	containerName := "Docker-web-server-example"
	envMap := map[string]any{
		"PORT": 80,
	} 
	portConfig := NewPortConfig("127.0.0.1", 80, 80, "tcp")

	ctx := context.Background()
	mobyClient, err := NewMobyClient(ctx)
	if err != nil {
		log.Fatal("encountered enough when creating moby api client", err)
	}

	containerConfig := NewContainerConfig(imageName, containerName, envMap, portConfig)

	containerlet := NewContainerlet(containerConfig, mobyClient)
	fmt.Println(containerlet)

}

// 	res, err := apiClient.ContainerCreate(ctx, client.ContainerCreateOptions{
// 		Name: containerName,
// 		Image: imageName,
// 		Config: &container.Config{
// 			Env: envVars,
// 			ExposedPorts: exposedPorts,
// 		},
// 		HostConfig: &container.HostConfig{
// 			PortBindings: portBindings,
// 		},
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Println("Container ID is:", res.ID)

// 	_, err = apiClient.ContainerStart(ctx, res.ID, client.ContainerStartOptions{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
	
// }
