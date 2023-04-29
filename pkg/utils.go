package pkg

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func CountDirectories(path string) int {
	dirCount := 0
	files, _ := os.ReadDir(path)
	for _, file := range files {
		if file.IsDir() {
			dirCount++
		}
	}
	return dirCount
}

func GetSubdirs(path string) ([]string, error) {
	subdirs := []string{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfos, err := file.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			subdirs = append(subdirs, filepath.Join(path, fileInfo.Name()))
		}
	}

	return subdirs, nil
}

func GetGroupFromMail(rootDir string, mail string) (string, error) {
	// Find the target directory
	matches, err := filepath.Glob(filepath.Join(rootDir, "**", mail))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("directory '%s' not found in '%s'", mail, rootDir)
	}

	// Get the parent directory path
	targetPath := matches[0]
	parentDir := filepath.Dir(targetPath)

	// Get the name of the parent directory
	_, parentDirName := filepath.Split(parentDir)

	return parentDirName, nil
}

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
func DirectoryExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func AddAdmin(name, fullname, mail string) error {
	admin := &Admin{Name: name, FullName: fullname, Mail: mail}
	err := DB.Create(admin).Error
	if err != nil {
		return err
	}
	return nil
}

func GetAdminByEmail(email string) (*Admin, error) {
	admin := &Admin{}
	err := DB.Where("mail = ?", email).First(admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return admin, nil
}

func AdminExists(email string) bool {
	var count int64
	DB.Model(&Admin{}).Where("mail = ?", email).Count(&count)
	return count > 0
}
func UserExists(email string) bool {
	var count int64
	DB.Model(&ClusterUser{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func AddUsers(mail, cluster, name, fullname string) error {
	user := ClusterUser{Name: name, FullName: fullname, Email: mail}
	if err := DB.Create(&user).Error; err != nil {
		return err
	}
	var c Cluster
	if err := DB.Where("name = ?", cluster).First(&c).Error; err != nil {
		return err
	}
	if err := DB.Model(&user).Association("Clusters").Append(&c); err != nil {
		return err
	}
	log.Println(fullname, "is added to the cluster", cluster)
	return nil
}

func GetUserByEmail(email string) (ClusterUser, error) {
	var user ClusterUser
	if err := DB.Where("email = ?", email).Preload("Clusters").First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func AddClusterToUser(email string, clusterName string) error {
	// Get the user with the given email
	var user ClusterUser
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		return err
	}

	// Get the cluster with the given name
	var cluster Cluster
	if err := DB.Where("name = ?", clusterName).First(&cluster).Error; err != nil {
		return err
	}

	// Add the cluster to the user's clusters
	if err := DB.Model(&user).Association("Clusters").Append(&cluster); err != nil {
		return err
	}
	log.Println("User with mail", email, "is added to the cluster named", clusterName)
	return nil
}

func CreateCluster(name, token string) error {
	// Create a new cluster instance
	cluster := Cluster{Name: name, Status: false, Token: token, API: "*"}

	// Save the cluster to the database
	err := DB.Create(&cluster).Error
	if err != nil {
		return err
	}

	return nil
}

func DeleteCluster(name string) error {
	// Find the cluster by name
	var cluster Cluster
	if err := DB.Where("name = ?", name).First(&cluster).Error; err != nil {
		return err
	}

	// Delete the cluster
	if err := DB.Delete(&cluster).Error; err != nil {
		return err
	}

	return nil
}

func ClusterExists(name string) bool {
	var c Cluster
	err := DB.Where("name = ?", name).First(&c).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Fatal(err)
	}
	if c.ID != 0 {
		return true
	}
	return false
}

func GetClusterByName(name string) (Cluster, error) {
	var cluster Cluster
	result := DB.Where("name = ?", name).First(&cluster)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Cluster{}, result.Error
		}
		return Cluster{}, result.Error
	}
	return cluster, nil
}

func GetAllClusters() ([]Cluster, error) {
	var clusters []Cluster
	result := DB.Find(&clusters)
	if result.Error != nil {
		return nil, result.Error
	}
	return clusters, nil
}

func UpdateCluster(name, api string) error {
	var cluster Cluster
	err := DB.Where("name = ?", name).First(&cluster).Error
	if err != nil {
		return err
	}

	cluster.Status = true
	cluster.API = api

	err = DB.Save(&cluster).Error
	if err != nil {
		return err
	}
	log.Println(name, "is activated and this is the api", api)
	return nil
}

func RemoveClusterFromUser(mail, clusterName string) error {
	var user ClusterUser
	err := DB.Where("email = ?", mail).Preload("Clusters", "name = ?", clusterName).First(&user).Error
	if err != nil {
		return err
	}

	if len(user.Clusters) == 0 {
		// User has no clusters left, delete user
		err = DB.Delete(&user).Error
		if err != nil {
			return err
		}
		log.Println("User", user.FullName, "deleted because they had no remaining clusters.")
		return nil
	}

	// Remove cluster from user's clusters
	var cluster Cluster
	err = DB.Where("name = ?", clusterName).First(&cluster).Error
	if err != nil {
		return err
	}
	association := DB.Model(&user).Association("Clusters")
	err = association.Delete(&cluster)
	if err != nil {
		return err
	}

	log.Println("Cluster", clusterName, "removed from user", user.FullName)
	return nil
}

func GetClustersByUser(email string) ([]Cluster, error) {
	var user ClusterUser
	if err := DB.Where("email = ?", email).Preload("Clusters").First(&user).Error; err != nil {
		return nil, err
	}
	return user.Clusters, nil
}

func GetUsersByCluster(clusterName string) ([]ClusterUser, error) {
	var users []ClusterUser
	if err := DB.Preload("Clusters", "name = ?", clusterName).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetAllUsers() ([]ClusterUser, error) {
	var users []ClusterUser
	if err := DB.Preload("Clusters").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
func UserInCluster(email, clusterName string) (bool, error) {
	var user ClusterUser
	err := DB.Preload("Clusters", "name = ?", clusterName).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	for _, cluster := range user.Clusters {
		if cluster.Name == clusterName {
			return true, nil
		}
	}
	return false, nil
}

func DeleteDir(root string, name string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == name {
			err := os.RemoveAll(path)
			if err != nil {
				return err
			}
			log.Printf("Directory %s deleted\n", path)
			return filepath.SkipDir
		}

		return nil
	})
}

func UploadFile(path string, r *http.Request) error {
	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Copy the uploaded file to the destination file
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}
	return nil

}
