package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"golang.org/x/term"
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

func pullImage(cli *client.Client, ctx context.Context) error {
	reader, err := cli.ImagePull(ctx, "osrf/ros:noetic-desktop-full", client.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(io.Discard, reader)
	fmt.Println("Pulling Image: osrf/ros:noetic-desktop-full")
	return nil
}

func RunContainer(cli *client.Client, ctx context.Context, codePath string, name string) {
	if check, err := imageExists(cli, ctx); err != nil {
		fmt.Println("Error contacting docker daemon:", err)
		return
	} else if !check {
		pullImage(cli, ctx)
	}
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
			Resources: container.Resources{
				Devices: []container.DeviceMapping{
					{
						PathOnHost:        "/dev/dri",
						PathInContainer:   "/dev/dri",
						CgroupPermissions: "rwm",
					},
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

func StartContainer(cli *client.Client, ctx context.Context, workspace string) error {
	if _, err := cli.ContainerStart(ctx, "ros-"+workspace, client.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

func imageExists(cli *client.Client, ctx context.Context) (bool, error) {
	filterArgs := client.Filters{}
	filterArgs.Add("reference", "osrf/ros:noetic-desktop-full")
	images, err := cli.ImageList(ctx, client.ImageListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return false, err
	}
	return len(images.Items) > 0, err
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

func EngageCommand(cli *client.Client, ctx context.Context, containerID string, command []string) error {
	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		TTY:          true,
	})
	if err != nil {
		return fmt.Errorf("exec Create error: %w", err)
	}
	attachResp, err := cli.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{
		TTY: true,
	})
	if err != nil {
		return fmt.Errorf("exec attach error: %w", err)
	}
	defer attachResp.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			attachResp.Conn.Write([]byte{3}) // Ctrl+C
			_ = sig
		}
	}()

	go io.Copy(attachResp.Conn, os.Stdin)
	_, err = io.Copy(os.Stdout, attachResp.Reader)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}
	for {
		inspect, err := cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
		if err != nil {
			return err
		}
		if !inspect.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func ExecIntoContainer(cli *client.Client, ctx context.Context, containerID string) error {
	execResp, err := cli.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		TTY:          true,
		Cmd:          []string{"/bin/bash"},
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
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	_, err = cli.ExecResize(ctx, execResp.ID, client.ExecResizeOptions{
		Height: uint(10 * height),
		Width:  uint(width),
	})
	if err != nil {
		return err
	}
	defer attachResp.Close()
	oldstate, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldstate)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// instead of handling sigs
	go io.Copy(os.Stdout, attachResp.Reader)
	go io.Copy(attachResp.Conn, os.Stdin)
	for {
		inspect, err := cli.ExecInspect(ctx, execResp.ID, client.ExecInspectOptions{})
		if err != nil {
			return err
		}
		if !inspect.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
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
		All:     true,
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

func GetWorkspaceMount(cli *client.Client, ctx context.Context, workspace string) (string, error) {
	inspRes, err := cli.ContainerInspect(ctx, "ros-"+workspace, client.ContainerInspectOptions{})
	if err != nil {
		return "", err
	}
	for _, mount := range inspRes.Container.Mounts {
		if mount.Destination == "/root/ros_ws" {
			return mount.Source, nil
		}
	}
	return "", fmt.Errorf("err: didn't find mounts on container (this should never have happened!!)") // Should never happen
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

func IsActive(cli *client.Client, ctx context.Context, contaienrID string) string {
	output, err := ExecCommand(cli, ctx, contaienrID, []string{
		"bash", "-c",
		"ps aux | grep rosmaster",
	})
	if err != nil {
		return "irresponsive"
	}
	if strings.Contains(output, "/opt/ros/noetic/bin/rosmaster") {
		return "active"
	} else {
		return "inactive"
	}
}
