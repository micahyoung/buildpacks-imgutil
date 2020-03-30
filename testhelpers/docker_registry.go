package testhelpers

import (
	"bytes"
	"context"
	"fmt"
	dockerconfigfile "github.com/docker/cli/cli/config/configfile"
	"golang.org/x/crypto/bcrypt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	dockerconfigtypes "github.com/docker/cli/cli/config/types"
	dockerapitypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	"github.com/docker/go-connections/nat"
)

var registryBaseImage = "registry:2"

type DockerRegistry struct {
	Name                string
	Host                string
	Port                string
	Username            string
	Password            string
	tmpDockerConfigPath string
}

func NewDockerRegistry() *DockerRegistry {
	return &DockerRegistry{
		Name:     "test-registry-" + RandString(10),
		Username: RandString(10),
		Password: RandString(10),
	}
}

func (registry *DockerRegistry) Start(t *testing.T) {
	t.Log("run registry")
	t.Helper()

	AssertNil(t, PullImage(DockerCli(t), registryBaseImage))
	ctx := context.Background()
	ctr, err := DockerCli(t).ContainerCreate(ctx, &container.Config{
		Image: registryBaseImage,
		Env: []string{
			"REGISTRY_STORAGE_DELETE_ENABLED=true",
			"REGISTRY_AUTH=htpasswd",
			"REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm",
			"REGISTRY_AUTH_HTPASSWD_PATH=/registry_test_htpasswd",
		},
	}, &container.HostConfig{
		AutoRemove: true,
		PortBindings: nat.PortMap{
			"5000/tcp": []nat.PortBinding{{}},
		},
	}, nil, registry.Name)
	AssertNil(t, err)


	htpasswdTar := generateHtpasswd(t, registry.Username, registry.Password)

	err = DockerCli(t).CopyToContainer(ctx, ctr.ID, "/", htpasswdTar, dockerapitypes.CopyToContainerOptions{})
	AssertNil(t, err)

	err = DockerCli(t).ContainerStart(ctx, ctr.ID, dockerapitypes.ContainerStartOptions{})
	AssertNil(t, err)

	inspect, err := DockerCli(t).ContainerInspect(ctx, ctr.ID)
	AssertNil(t, err)
	registry.Port = inspect.NetworkSettings.Ports["5000/tcp"][0].HostPort

	registry.Host, err = getRegistryHostname()
	AssertNil(t, err)

	registry.tmpDockerConfigPath = setDockerConfigEnvForDefaultKeychain(t, registry.Host, registry.Port, registry.Username, registry.Password)
	output, err := ioutil.ReadFile(registry.tmpDockerConfigPath)
	AssertNil(t, err)

	fmt.Printf("CREDS: %s %s", registry.Username, registry.Password)
	fmt.Printf("CONFIG: %s", string(output))
	Eventually(t, func() bool {
		txt, err := HttpGetAuthE(fmt.Sprintf("http://%s:%s/v2/", registry.Host, registry.Port), registry.Username, registry.Password)
		return err == nil && txt != ""
	}, 100*time.Millisecond, 10*time.Second)
}

func (registry *DockerRegistry) Stop(t *testing.T) {
	t.Log("stop registry")
	t.Helper()
	if registry.Name != "" {
		DockerCli(t).ContainerKill(context.Background(), registry.Name, "SIGKILL")
		DockerCli(t).ContainerRemove(context.TODO(), registry.Name, dockerapitypes.ContainerRemoveOptions{Force: true})
	}

	err := os.Remove(registry.tmpDockerConfigPath)
	AssertNil(t, err)
}

func (registry *DockerRegistry) NewTestImageName(suffixOpt ...string) string {
	suffix := ""
	if len(suffixOpt) == 1 {
		suffix = suffixOpt[0]
	}
	return fmt.Sprintf("%s:%s/pack-image-test-%s%s", registry.Host, registry.Port, RandString(10), suffix)
}

func setDockerConfigEnvForDefaultKeychain(t *testing.T, hostname, port, username, password string) string {
	tmpConfigDir, err := ioutil.TempDir("", "docker-config-default-keychain")
	AssertNil(t, err)

	tmpConfigPath := filepath.Join(tmpConfigDir, "config.json")

	config := dockerconfigfile.New(tmpConfigPath)
	config.AuthConfigs[hostname+":"+port] = dockerconfigtypes.AuthConfig{
		Username: username,
		Password: password,
	}

	err = config.Save()
	AssertNil(t, err)

	err = os.Setenv("DOCKER_CONFIG", tmpConfigDir)
	AssertNil(t, err)

	return tmpConfigPath
}

func generateHtpasswd(t *testing.T, username string, password string) io.Reader {
	var err error

	//https://docs.docker.com/registry/deploying/#restricting-access
	//https://github.com/foomo/htpasswd/blob/e3a90e78da9cff06a83a78861847aa9092cbebdd/hashing.go#L23
	passwordBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tarPath, err := CreateSingleFileTar("/registry_test_htpasswd", username+":"+string(passwordBytes))
	AssertNil(t, err)
	defer os.Remove(tarPath)

	tarBytes, err := ioutil.ReadFile(tarPath)
	AssertNil(t, err)

	err = os.Remove(tarPath)
	AssertNil(t, err)

	return bytes.NewBuffer(tarBytes)
}

func getRegistryHostname() (string, error) {
	dockerHost := os.Getenv(("DOCKER_HOST"))
	if dockerHost != "" {
		url, err := url.Parse(dockerHost)
		if err != nil {
			return "", err
		}
		return url.Hostname(), nil
	}
	return "localhost", nil
}
