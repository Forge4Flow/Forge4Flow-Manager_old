package models

import (
	docker_utils "github.com/forge4flow/forge4flow-manager/internal/docker"
	"gorm.io/gorm"
)

type Database struct {
	gorm.Model      `json:"-"`
	CoreInstanceID  uint                `gorm:"not null" json:"-"`
	Type            docker_utils.DbType `gorm:"not null" json:"type"`
	ContainerID     string              `json:"containerID,omitempty"`
	ContainerHealth string              `json:"containerHealth,omitempty"` // Healthy, Unhealthy, Starting, Stopped
	ContainerStatus string              `json:"containerStatus,omitempty"` // Pending, Creating, Running, Paused, Restarting, Removing, Exited, Dead
}
