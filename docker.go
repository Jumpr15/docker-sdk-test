package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"

	// "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type ContainerLifecycle interface {
	CreateContainer()
	StopContainer()
}

type Containerlet struct {
	ImageName string
	ContainerName string
	EnvMap map[string]any
	PortConfig PortConfig

	ContainerClient ContainerLifecycle

	ContainerID string // set manually or via container ID
}

func NewContainerlet(imageName string, containerName string, EnvMap map[string]any, PortConfig PortConfig, ContainerClient ContainerLifecycle) (*Containerlet) {
	client := Containerlet{
		ImageName: imageName,
		ContainerName: containerName,
		EnvMap: EnvMap,
		PortConfig: PortConfig,
		ContainerClient: ContainerClient,
	}
	return &client
}

type MobyClient struct {
	Ctx context.Context 
	Client *client.Client
}

func (mb *MobyClient) CreateContainer() {}

func (mb *MobyClient) StopContainer() {}

func NewMobyClient(ctx context.Context) (*MobyClient) {
	mobyClient := MobyClient{
		Ctx: ctx,
	}
	return &mobyClient
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
	// envVars := MapToEnvArr(envMap)

	portConfig := NewPortConfig("127.0.0.1", 80, 80, "tcp")
	// portConfigList := []PortConfig{portConfig}

	// exposedPorts, portBindings := ConvertPortConfig(portConfigList)

	//

	ctx := context.Background()
	mobyClient := NewMobyClient(ctx)

	//

	containerlet := NewContainerlet(imageName, containerName, envMap, portConfig,mobyClient)
	fmt.Println(containerlet)

}
	// apiClient, err := client.New(client.FromEnv)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer apiClient.Close()



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
