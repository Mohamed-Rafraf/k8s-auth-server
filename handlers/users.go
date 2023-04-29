package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mohamed-rafraf/k8s-auth-server/pkg"
)

func HandleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		HandlePostUser(w, r)
		return
	case http.MethodDelete:
		HandleDeleteUser(w, r)
	case http.MethodGet:
		HandleGetUser(w, r)
		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"status": "error", "message": "Method not allowed"}`))
	}
}

func HandlePostUser(w http.ResponseWriter, r *http.Request) {
	//var user pkg.ClusterUser
	var err error
	var group_exist bool
	var user_exist bool
	var user_and_cluster bool

	userType, email, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(email, " not Authorized to create user")
		res = pkg.Response{
			Status:  "fail",
			Message: msg,
		}
		err = encoder.Encode(res)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	name := r.FormValue("name")
	fullname := r.FormValue("fullname")
	cluster := r.FormValue("cluster")
	mail := r.FormValue("mail")
	group := r.FormValue("group")
	cluster_exist := pkg.ClusterExists(cluster)
	if name == "" {
		res = pkg.Response{
			Status:  "fail",
			Message: "Please specify the name",
		}
		goto SEND
	}
	if fullname == "" {
		res = pkg.Response{
			Status:  "fail",
			Message: "Please specify the fullname",
		}
		goto SEND
	}
	if cluster == "" {
		res = pkg.Response{
			Status:  "fail",
			Message: "Please specify the cluster",
		}
		goto SEND
	}

	if err != nil {
		log.Println(err)
		return
	}
	if !cluster_exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "The cluster is not exist!",
		}
		goto SEND
	}
	if mail == "" {
		res = pkg.Response{
			Status:  "fail",
			Message: "Please specify the mail",
		}
		goto SEND
	}

	user_and_cluster, _ = pkg.UserInCluster(mail, cluster)

	if user_and_cluster {
		res = pkg.Response{
			Status:  "fail",
			Message: "This user is exist in this cluster",
		}
		goto SEND
	}

	user_exist = pkg.UserExists(mail)

	if !user_and_cluster && user_exist {
		err = pkg.AddClusterToUser(mail, cluster)
		if err != nil {
			panic(err)
		}
	}
	if !user_exist {
		err = pkg.AddUsers(mail, cluster, name, fullname)
		if err != nil {
			panic(err)
		}
	}
	if group == "" {
		// Get the uploaded file from the form data
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		err = os.Mkdir("clusters/"+cluster+"/"+mail, 0755)
		if err != nil {
			panic(err)
		}

		f, err := os.Create("clusters/" + cluster + "/" + mail + "/RBAC.yaml")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		// Copy the uploaded file to the destination file
		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		group_exist = pkg.DirectoryExists("clusters/" + cluster + "/" + group)
		if !group_exist {
			res = pkg.Response{
				Status:  "fail",
				Message: "this group didn't exist",
			}
			goto SEND
		}
		err = os.Mkdir("clusters/"+cluster+"/"+group+"/"+mail, 0755)
		if err != nil {
			panic(err)
		}

	}
	log.Println(email, "create new user with mail", mail, "in ", cluster)
	res = pkg.Response{
		Status:  "success",
		Message: "The user with mail " + mail + " is created in this cluster " + cluster,
	}
SEND:
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}
}

func HandleGetUser(w http.ResponseWriter, r *http.Request) {
	userType, email, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.UserResponse
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(email, "is not authorized to see Users")
		res = pkg.UserResponse{
			Status:       "fail",
			Message:      msg,
			ClusterUsers: []pkg.ClusterUser{},
		}
		err := encoder.Encode(res)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}
	users, err := pkg.GetAllUsers()

	if err != nil {
		log.Println(err)
		return
	}

	res = pkg.UserResponse{
		Status:       "success",
		Message:      "All users with their clusters",
		ClusterUsers: users,
	}
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {

	var err error
	var user_exist bool

	userType, email, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(email, "are not authorized to delete user")
		res = pkg.Response{
			Status:  "fail",
			Message: msg,
		}
		err = encoder.Encode(res)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}
	cluster := r.FormValue("cluster")
	mail := r.FormValue("mail")
	cluster_exist := pkg.ClusterExists(cluster)
	if cluster == "" {
		res = pkg.Response{
			Status:  "fail",
			Message: "Please specify the cluster",
		}
		goto SEND
	}

	if !cluster_exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "The cluster is not exist!",
		}
		goto SEND
	}

	if mail == "" {
		res = pkg.Response{
			Status:  "fail",
			Message: "Please specify the mail",
		}
		goto SEND
	}

	user_exist = pkg.UserExists(mail)

	if !user_exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "Used don't exist",
		}
		goto SEND
	}
	err = pkg.RemoveClusterFromUser(mail, cluster)
	if err != nil {
		log.Println(err)
		return
	}
	err = pkg.DeleteDir("clusters/"+cluster, mail)
	if err != nil {
		log.Println(err)
		return
	}
	res = pkg.Response{
		Status:  "success",
		Message: "The user deleted successfully!",
	}
	_, err = SendMessageToClient(cluster, "Token-Delete,user="+mail)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(email, "delete", mail, "from", cluster)
SEND:
	err = encoder.Encode(res)
	fmt.Println(res)
	if err != nil {
		log.Println(err)
		return
	}

}
