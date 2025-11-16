package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

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
			Env:       []string{"DISPLAY=" + os.Getenv("DISPLAY")},
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

func ExecBackgroundCommand(cli *client.Client, ctx context.Context, containerID string, command []string) error {
	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          command,
		AttachStdout: false,
		AttachStderr: false,
		TTY:          false,
	})
	if err != nil {
		return fmt.Errorf("exec Create error: %w", err)
	}
	_, err = cli.ExecStart(ctx, execResp.ID, client.ExecStartOptions{
		Detach: true,
		TTY:    false,
	})
	if err != nil {
		return fmt.Errorf("exec start error: %w", err)
	}
	return nil
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

func ExecIntoContainer(cli *client.Client, ctx context.Context, containerID string) error {
	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
		Cmd:          []string{"/bin/bash", "-c", "stty -echo; exec bash"},
	})
	if err != nil {
		return err
	}
	attachResp, err := cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{
		TTY: true,
	})
	if err != nil {
		return err
	}
	defer attachResp.Close()
	go io.Copy(os.Stdout, attachResp.Reader)
	go func() {
		io.Copy(attachResp.Conn, os.Stdin)
		attachResp.CloseWrite()
	}()
	for {
		inspectResp, err := cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
		if err != nil {
			return err
		}
		if !inspectResp.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
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
		if container.Names[0][1:] == name {
			return true, nil
		}
	}
	return false, nil
}

func ListRunningContainers(cli *client.Client, ctx context.Context) ([]string, error) {
	filterArgs := client.Filters{}
	filterArgs.Add("ancestor", "osrf/ros:noetic-desktop-full")

	containers, err := cli.ContainerList(ctx, client.ContainerListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}
	var workspaces []string
	for _, container := range containers.Items {
		if strings.HasPrefix(container.Names[0][1:], "ros-") {
			workspaces = append(workspaces, container.Names[0][5:])
		}
	}
	return workspaces, nil
}

func StopAndDeleteContainer(cli *client.Client, ctx context.Context, containerID string) error {
	timeout := 5
	_, err := cli.ContainerStop(ctx, containerID, client.ContainerStopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return err
	}
	_, err = cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		return err
	}
	return nil
}
