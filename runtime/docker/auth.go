package docker

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
)

func GetImageDomainName(imageName string) string {
	imageRef, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		panic(err)
	}

	domainName := reference.Domain(imageRef)

	return domainName
}

func GetDockerConfigPath(configPath string) (string, error) {
	var dockerConfigError error
	var dockerConfigPath string

	if configPath == "" {
		homeDir, err := os.UserHomeDir()

		if err != nil {
			dockerConfigError = err
		}

		dockerConfigPath = path.Join(homeDir, ".docker", "config.json")
	} else {
		dockerConfigPath = configPath
	}

	return dockerConfigPath, dockerConfigError
}

func GetDockerConfig(configPath string) (*DockerConfig, error) {
	dockerConfigPath, err := GetDockerConfigPath(configPath)

	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(dockerConfigPath)
	if err != nil {
		log.Errorf("Unable to read docker config, error: %v", err)
		return nil, err
	}

	var dockerConfig DockerConfig
	jsonError := json.Unmarshal(file, &dockerConfig)

	if jsonError != nil {
		log.Errorf("Unable to Unmarshal docker config, error: %v", jsonError)
		return nil, jsonError
	}

	return &dockerConfig, nil
}

func GetDockerAuth(dockerConfig *DockerConfig, imageName string) (string, error) {
	imageDomain := GetImageDomainName(imageName)

	if authString, ok := dockerConfig.Auths[imageDomain]["auth"]; ok {
		decodedAuth, err := base64.URLEncoding.DecodeString(authString)
		if err != nil {
			return "", err
		}

		decodedAuthSplit := strings.Split(string(decodedAuth), ":")
		authConfig := types.AuthConfig{
			Username: decodedAuthSplit[0],
			Password: decodedAuthSplit[1],
		}

		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return "", err
		}

		authString := base64.URLEncoding.EncodeToString(encodedJSON)
		return authString, nil
	}

	return "", nil
}
