package ExternalMetrics

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type DockerContainer struct {
	ID    string
	Name  []string
	State string
}

func ListDockerContainers() []DockerContainer {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	err_list := make([]DockerContainer, 0)
	if err != nil {
		return err_list
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err_list
	}
	list := make([]DockerContainer, len(containers))
	for i, container := range containers {
		list[i].ID = container.ID
		list[i].Name = container.Names
		list[i].State = container.State
	}
	return list
}
