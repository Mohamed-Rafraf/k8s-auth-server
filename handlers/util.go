package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/mohamed-rafraf/k8s-auth-server/pkg"
)

func HandlePermissions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		HandleGetPermissions(w, r)
	case http.MethodPost:
		HandlePostPermissions(w, r)
		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"status": "error", "message": "Method not allowed"}`))
	}
}

func HandleGetPermissions(w http.ResponseWriter, r *http.Request) {
	var rbacFile string
	var content []byte

	userType, email, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(email, "is not authorized to see permissions")
		res = pkg.Response{
			Status:  "fail",
			Message: msg,
		}
		err := encoder.Encode(res)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}
	typeP := r.FormValue("type")
	name := r.FormValue("name")
	cluster := r.FormValue("cluster")
	cluster_exist := pkg.ClusterExists(cluster)

	var cluster_path string

	if !cluster_exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "The cluster is not exist!",
		}
		goto SEND
	}
	cluster_path = "clusters/" + cluster + "/"
	if typeP == "user" {
		if !pkg.UserExists(name) {
			res = pkg.Response{
				Status:  "fail",
				Message: "User don't exist",
			}
			goto SEND

		}
		group, _ := pkg.GetGroupFromMail(cluster_path, name)
		if group == "" {
			rbacFile = cluster_path + name + "/RBAC.yaml"
		} else {
			rbacFile = cluster_path + group + "/RBAC.yaml"
		}
		content, err = os.ReadFile(rbacFile)

		if err != nil {
			res = pkg.Response{
				Status:  "fail",
				Message: "Failed to read RBAC file ",
			}
			log.Println(err)
			goto SEND
		}
	} else if typeP == "group" {
		rbacFile = cluster_path + name + "/RBAC.yaml"
		content, err = os.ReadFile(rbacFile)
		if err != nil {
			res = pkg.Response{
				Status:  "fail",
				Message: "Failed to read RBAC file ",
			}
			log.Println(err)
			goto SEND
		}

	}

	rbacFile = string(content)
	res = pkg.Response{
		Status:  "success",
		Message: rbacFile,
	}
SEND:
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandlePostPermissions(w http.ResponseWriter, r *http.Request) {
	userType, email, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(email, "is not authorized to give permissions")
		res = pkg.Response{
			Status:  "fail",
			Message: msg,
		}
		err := encoder.Encode(res)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}
	typeP := r.FormValue("type")
	name := r.FormValue("name")
	cluster := r.FormValue("cluster")
	cluster_exist := pkg.ClusterExists(cluster)

	var cluster_path string

	if !cluster_exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "The cluster is not exist!",
		}
		goto SEND
	}

	cluster_path = "clusters/" + cluster + "/"
	if typeP == "user" {
		if !pkg.UserExists(name) {
			res = pkg.Response{
				Status:  "fail",
				Message: "User don't exist",
			}
			goto SEND

		}
		group, _ := pkg.GetGroupFromMail(cluster_path, name)
		if group != "" {
			_ = os.RemoveAll(cluster_path + group + "/" + name)
			_ = os.Mkdir(cluster_path+name, 0755)
		}
		err = pkg.UploadFile("clusters/"+cluster+"/"+name+"/RBAC.yaml", r)
		if err != nil {
			panic(err)
		}

	} else if typeP == "group" {
		err = pkg.UploadFile("clusters/"+cluster+"/"+name+"/RBAC.yaml", r)
		if err != nil {
			panic(err)
		}

	}
	res = pkg.Response{
		Status:  "success",
		Message: "Permission Upadted",
	}
SEND:
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}
