package docker_utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func IsDockerInstalled() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

func InstallDocker() error {
	fmt.Println("Installing Docker...")
	switch runtime.GOOS {
	case "windows":
		err := installDockerWindows()
		if err != nil {
			return err
		}
	case "darwin":
		err := installDockerMacOS()
		if err != nil {
			return err
		}
	case "linux":
		err := installDockerLinux()
		if err != nil {
			return err
		}
	default:
		fmt.Println("Unsupported operating system:", runtime.GOOS)
		os.Exit(1)

	}
	fmt.Println("Docker installation completed.")
	return nil
}

func installDockerWindows() error {
	cmd := exec.Command("powershell", "-Command", "Invoke-WebRequest -Uri https://get.docker.com/builds/Windows/x86_64/docker-latest.zip -OutFile $HOME/Downloads/docker-latest.zip")
	err := cmd.Run()
	if err != nil {
		return err
	}
	// Extract the downloaded ZIP file and install Docker
	// Write the necessary code here to extract and install Docker on Windows
	return nil
}

func installDockerMacOS() error {
	cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func installDockerLinux() error {
	cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
