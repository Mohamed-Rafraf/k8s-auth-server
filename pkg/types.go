package pkg

import (
	"gorm.io/gorm"
)

type UserCluster struct {
	gorm.Model
	ClusterID uint
	UserID    uint
}

type ClusterUser struct {
	gorm.Model
	Name     string    `gorm:"not null"`
	FullName string    `gorm:"not null"`
	Email    string    `gorm:"unique;not null"`
	Clusters []Cluster `gorm:"many2many:user_clusters;"`
}

type Admin struct {
	gorm.Model
	Name     string
	Mail     string `gorm:"unique"`
	FullName string
}

type Cluster struct {
	gorm.Model
	Name   string `gorm:";not null"`
	Status bool
	Token  string
	API    string
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

type UserResponse struct {
	Status       string        `json:"status"`
	Message      string        `json:"message"`
	ClusterUsers []ClusterUser `json:"clusterusers"`
}

type ClusterResponse struct {
	Status   string    `json:"status"`
	Message  string    `json:"message"`
	Clusters []Cluster `json:"clusters"`
}
