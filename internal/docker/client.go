package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
)

var (
	cli  *client.Client
	once sync.Once
)

func GetClient() (*client.Client, context.Context, error) {
	var err error
	once.Do(func() {
		cli, err = client.New(client.FromEnv, client.WithAPIVersionNegotiation())
	})
	if err != nil {
		return nil, nil, err
	}
	ctx := context.Background()
	return cli, ctx, nil
}

func RunContainer(cli *client.Client, ctx context.Context, codePath string, name string) {
	resp, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Name: "ros-" + name,
		Config: &container.Config{
			Image:     "osrf/ros:noetic-desktop-full",
			Tty:       true,
			OpenStdin: true,
		},
		HostConfig: &container.HostConfig{
			NetworkMode: container.NetworkMode("host"),
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/tmp/.X11-unix",
					Target: "/tmp/.X11-unix",
				},
				{
					Type:   mount.TypeBind,
					Source: codePath,
					Target: "/root/ros_ws",
				},
			},
		},
	})
	if err != nil {
		fmt.Println("Error creating container:", err)
		return
	}
	if _, err := cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		fmt.Println("Error starting container:", err)
		return
	}
}

func ExecCommand(cli *client.Client, ctx context.Context, containerID string, command []string) (string, error) {
	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          false,
	})
	if err != nil {
		return "", fmt.Errorf("exec Create error: %w", err)
	}
	attachResp, err := cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("exec attach error: %w", err)
	}
	defer attachResp.Close()

	var outBuf bytes.Buffer
	_, err = io.Copy(&outBuf, attachResp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}
	return outBuf.String(), nil
}

func SearchRunningContainers(name string) (bool, error) {
	cli, ctx, err := GetClient()
	if err != nil {
		return false, err
	}
	containers, err := cli.ContainerList(ctx, client.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return false, err
	}
	for _, container := range containers.Items {
		if container.Names[0] == name {
			return true, nil
		}
	}
	return false, nil
}
