package docker_utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DbType string

const (
	CoreDB   DbType = "Core"
	EventsDB DbType = "CoreEvents"
)

type Config struct {
	Name      string
	NetworkId string
	Database  DatabaseConfig
	Core      CoreConfig
}

type DatabaseConfig struct {
	User     string
	Password string
}

type CoreConfig struct {
	ManagerInstall bool
	Port           int64
	AutoMigrate    bool
	ApiKey         string
	FlowNetwork    string
	AdminAccount   string
}

func DeployForge4FlowContainers(cfg *Config) error {
	cfg, err := createNetwork(cfg)
	if err != nil {
		return err
	}

	err = createDatabaseContainer(cfg, CoreDB)
	if err != nil {
		return err
	}

	err = createDatabaseContainer(cfg, EventsDB)
	if err != nil {
		return err
	}

	err = createCoreContainer(cfg)
	if err != nil {
		return err
	}

	return nil
}

func getClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return cli
}

func createNetwork(cfg *Config) (*Config, error) {
	ctx := context.Background()

	cli := getClient()

	// Create the network
	response, err := cli.NetworkCreate(ctx, cfg.Name+"-network", types.NetworkCreate{})
	if err != nil {
		return nil, err
	}

	cfg.NetworkId = response.ID

	return cfg, nil
}

func createDatabaseContainer(cfg *Config, dbType DbType) error {
	ctx := context.Background()

	cli := getClient()

	// Pull Docker Image
	imageName := "mysql:8.0.32"
	err := pullImage(cli, imageName)
	if err != nil {
		return err
	}

	// Database container
	databaseContainer, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Env: []string{
			"MYSQL_USER=" + cfg.Database.User,
			"MYSQL_PASSWORD=" + cfg.Database.Password,
			"MYSQL_DATABASE=" + cfg.Name + string(dbType),
			"MYSQL_RANDOM_ROOT_PASSWORD=true",
		},
		Healthcheck: &container.HealthConfig{
			Test:     []string{"CMD", "mysqladmin", "ping", "-h", "localhost"},
			Interval: time.Duration(5) * time.Second,
			Retries:  10,
		},
	}, nil, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			cfg.Name + "-network": {
				NetworkID: cfg.NetworkId,
			},
		},
	}, nil, cfg.Name+string(dbType)+"-database")
	if err != nil {
		return err
	}

	// Starting the database container
	if err := cli.ContainerStart(ctx, databaseContainer.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Wait for the database container to be healthy
	_, err = waitForHealthyContainer(cli, databaseContainer.ID)
	if err != nil {
		return err
	}

	return nil
}

func createCoreContainer(cfg *Config) error {
	ctx := context.Background()

	cli := getClient()

	// Pull Docker Image
	imageName := "forge4flow/forge4flow-core:0.0.1"
	err := pullImage(cli, imageName)
	if err != nil {
		return err
	}

	// Core container
	coreContainer, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Env: []string{
			"FORGE4FLOW_MANAGERINSTALL=" + fmt.Sprint(cfg.Core.ManagerInstall),
			"FORGE4FLOW_FLOWNETWORK=" + fmt.Sprint(cfg.Core.FlowNetwork),
			"FORGE4FLOW_ADMINACCOUNT=" + fmt.Sprint(cfg.Core.AdminAccount),
			"FORGE4FLOW_AUTOMIGRATE=" + fmt.Sprint(cfg.Core.AutoMigrate),
			"FORGE4FLOW_AUTHENTICATION_APIKEY=" + cfg.Core.ApiKey,
			"FORGE4FLOW_DATASTORE_MYSQL_USERNAME=" + cfg.Database.User,
			"FORGE4FLOW_DATASTORE_MYSQL_PASSWORD=" + cfg.Database.Password,
			"FORGE4FLOW_DATASTORE_MYSQL_HOSTNAME=" + cfg.Name + string(CoreDB) + "-database",
			"FORGE4FLOW_DATASTORE_MYSQL_DATABASE=" + cfg.Name + string(CoreDB),
			"FORGE4FLOW_EVENTSTORE_MYSQL_USERNAME=" + cfg.Database.User,
			"FORGE4FLOW_EVENTSTORE_MYSQL_PASSWORD=" + cfg.Database.Password,
			"FORGE4FLOW_EVENTSTORE_MYSQL_HOSTNAME=" + cfg.Name + string(EventsDB) + "-database",
			"FORGE4FLOW_EVENTSTORE_MYSQL_DATABASE=" + cfg.Name + string(EventsDB),
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"8000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprint(cfg.Core.Port),
				},
			},
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			cfg.Name + "-network": {
				NetworkID: cfg.NetworkId,
			},
		},
	}, nil, cfg.Name+"-core")
	if err != nil {
		return err
	}

	// Starting the container
	if err := cli.ContainerStart(ctx, coreContainer.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// TODO: Fix Health Check
	// // Wait for the container to be healthy
	// _, err = waitForHealthyContainer(cli, coreContainer.ID)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func pullImage(cli *client.Client, imageName string) error {
	resp, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer resp.Close()
	io.Copy(os.Stdout, resp)

	return nil
}

func waitForHealthyContainer(cli *client.Client, containerID string) (*types.ContainerJSON, error) {
	ctx := context.Background()
	timeout := 2 * time.Minute
	start := time.Now()

	for {
		// Check if the timeout has been exceeded
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timed out waiting for container to become healthy")
		}

		// Use the ContainerInspect method to get the container status and health
		containerJSON, err := cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return nil, err
		}

		if containerJSON.State.Health != nil && containerJSON.State.Health.Status == "healthy" {
			return &containerJSON, nil
		}

		// Wait for a short period before checking again
		time.Sleep(1 * time.Second)
	}
}

func GenerateRandomAPIKey() (string, error) {
	keyLength := 32

	// Calculate the number of bytes required to generate the API key
	byteLength := keyLength / 4 * 3 // 4 base64 characters represent 3 bytes

	if keyLength%4 != 0 {
		byteLength += 3
	}

	// Generate random bytes
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode the random bytes to base64
	apiKey := base64.URLEncoding.EncodeToString(bytes)

	// Trim any trailing padding characters ('=')
	apiKey = apiKey[:keyLength]

	return apiKey, nil
}
