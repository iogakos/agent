// Copyright 2022 Metrika Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dt "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

var (
	// ErrContainerNotFound container not found
	ErrContainerNotFound = errors.New("container not found")

	// ErrEmptyLogFile log file is empty
	ErrEmptyLogFile = errors.New("log file is empty")

	// DefaultDockerHost host docker daemon address to connect to
	DefaultDockerHost = ""

	// DefaultDockerAdapter default docker adapter for container discovery.
	DefaultDockerAdapter = DockerAdapter(&DockerProductionAdapter{})
)

// DockerAdapter container discovery interface.
type DockerAdapter interface {
	// GetRunningContainers returns a slice of all
	// currently running Docker containers
	GetRunningContainers() ([]dt.Container, error)

	// MatchContainer takes a slice of containers and regex strings.
	// It returns the first running container to match any of the identifiers.
	// If no matches are found, ErrContainerNotFound is returned.
	MatchContainer(containers []dt.Container, identifiers []string) (dt.Container, error)

	DockerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)

	DockerEvents(ctx context.Context, options types.EventsOptions) (
		<-chan events.Message, <-chan error, error)
}

// DockerProductionAdapter adapter for accessing the host docker daemon
type DockerProductionAdapter struct{}

// GetRunningContainers returns a slice of all
// currently running Docker containers
func (a *DockerProductionAdapter) GetRunningContainers() ([]dt.Container, error) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containers, err := cli.ContainerList(ctx, dt.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

// MatchContainer takes a slice of containers and regex strings.
// It returns the first running container to match any of the identifiers.
// If no matches are found, ErrContainerNotFound is returned.
func (a *DockerProductionAdapter) MatchContainer(containers []dt.Container, identifiers []string) (dt.Container, error) {
	for _, container := range containers {
		for _, rStr := range identifiers {
			r, err := regexp.Compile(rStr)
			if err != nil {
				return dt.Container{}, err
			}

			// Try to match the identifier with container names
			for _, name := range container.Names {
				if r.MatchString(name) {
					return container, nil
				}
			}
			// Try to match the identifier with Image name
			if r.MatchString(container.Image) {
				return container, nil
			}

		}
	}
	return dt.Container{}, ErrContainerNotFound
}

// DockerLogs returns a container's logs
func (a *DockerProductionAdapter) DockerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}

	return cli.ContainerLogs(ctx, container, options)
}

// DockerEvents gets channels for consuming docker events subscription messages and errors
func (a *DockerProductionAdapter) DockerEvents(ctx context.Context, options types.EventsOptions) (
	<-chan events.Message, <-chan error, error,
) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, nil, err
	}

	msgchan, errchan := cli.Events(ctx, options)
	return msgchan, errchan, nil
}

// GetRunningContainers convenience wrapper to the default adapter for
// getting running containers.
func GetRunningContainers() ([]dt.Container, error) {
	return DefaultDockerAdapter.GetRunningContainers()
}

// MatchContainer convenience wrapper for finding containers using the
// default adapter.
func MatchContainer(containers []dt.Container, identifiers []string) (dt.Container, error) {
	return DefaultDockerAdapter.MatchContainer(containers, identifiers)
}

// DockerLogs convenience wrapper for reading container logs.
func DockerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return DefaultDockerAdapter.DockerLogs(ctx, container, options)
}

// DockerEvents convenience wrapper for subscribing to docker events.
func DockerEvents(ctx context.Context, options types.EventsOptions) (
	<-chan events.Message, <-chan error, error,
) {
	return DefaultDockerAdapter.DockerEvents(ctx, options)
}

// PidOf returns the PID of a specified process name.
// If process is not found an os.ExitError is returned.
// In case of abnormal output, strconv.NumError is returned.
func PidOf(name string) (int, error) {
	var err error
	var ret int
	// try pidof
	output, err := exec.Command("pidof", "-s", name).Output()
	if err != nil {
		output, err = exec.Command("pgrep", "-n", name).Output()
		if err != nil {
			return 0, err
		}
	}

	ret, err = strconv.Atoi(strings.Trim(string(output), "\n"))
	if err != nil {
		return 0, err
	}
	return ret, nil
}

// PidArgs returns a string slice of the command line arguments
// of a specifid PID.
// First element is the executable path.
func PidArgs(pid int) ([]string, error) {
	pidStr := strconv.Itoa(pid)

	out, err := ioutil.ReadFile("/proc/" + pidStr + "/cmdline")
	if err != nil {
		return nil, err
	}
	args := bytes.Replace(out, []byte{0x0}, []byte{' '}, -1)
	return strings.Fields(string(args)), nil
}

// GetEnvFromFile returns a map of environment variables parsed from a file.
func GetEnvFromFile(path string) (map[string]string, error) {
	return godotenv.Read(path)
}

// GetLogLine wraps a reader and returns the first line of text.
// Use to determine the validity of the log file.
func GetLogLine(r io.Reader) ([]byte, error) {
	scan := bufio.NewScanner(r)
	ok := scan.Scan()
	if !ok {
		err := scan.Err()
		if err != nil {
			return nil, scan.Err()
		}
		return nil, ErrEmptyLogFile
	}
	return scan.Bytes(), nil
}

var dockerCLI *client.Client

func getDockerClient() (*client.Client, error) {
	if dockerCLI != nil {
		return dockerCLI, nil
	}

	defaultOpts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}

	if DefaultDockerHost != "" {
		defaultOpts = append(defaultOpts, client.WithHTTPClient(
			&http.Client{
				Transport: &http.Transport{
					Dial: func(network, addr string) (net.Conn, error) {
						return net.DialTimeout(network, addr, time.Second)
					},
				},
			}))
	}

	var err error
	dockerCLI, err = client.NewClientWithOpts(defaultOpts...)
	if err != nil {
		return nil, err
	}

	return dockerCLI, nil
}
