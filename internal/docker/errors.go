package docker_utils

type DockerError struct {
	message string
}

func (de DockerError) Error() string {
	return de.message
}

var (
	ErrInvalidInput = DockerError{"Invalid input"}
	ErrNotFound     = DockerError{"Item not found"}
	ErrPermission   = DockerError{"Permission denied"}
)
