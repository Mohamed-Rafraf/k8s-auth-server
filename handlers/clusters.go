package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/mohamed-rafraf/k8s-auth-server/pkg"
)

func HandleClusters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		HandlePostCluster(w, r)
	case http.MethodDelete:
		HandleDeleteCluster(w, r)
	case http.MethodGet:
		HandleGetClusters(w, r)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"status": "error", "message": "Method not allowed"}`))
	}
}

func HandlePostCluster(w http.ResponseWriter, r *http.Request) {
	userType, email, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(email, "try to create a cluster but he's not an admin")
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
	name := r.FormValue("name")
	cluster := pkg.ClusterExists(name)

	if cluster {
		res = pkg.Response{
			Status:  "fail",
			Message: "Cluster is already registred!",
		}
		log.Println(cluster, "is already registred")
	} else {
		token, err := pkg.GenerateToken(20)
		if err != nil {
			log.Println(err)
			return
		}
		err = pkg.CreateCluster(name, token)
		if err != nil {
			log.Println(err)
			return
		}

		res = pkg.Response{
			Status:  "success",
			Message: token,
		}
		log.Println(name, "is create successfully by", email)
		err = os.Mkdir("clusters/"+name, 0755)
		if err != nil {
			panic(err)
		}
	}
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandleGetClusters(w http.ResponseWriter, r *http.Request) {

	userType, email, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
	}
	var res pkg.ClusterResponse
	encoder := json.NewEncoder(w)
	if userType == "none" {
		log.Println(email, "try to list clusters but he's not authorized")
		res = pkg.ClusterResponse{
			Status:   "fail",
			Message:  msg,
			Clusters: []pkg.Cluster{},
		}
		err := encoder.Encode(res)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	var clusters []pkg.Cluster
	if userType == "user" {
		clusters, err = pkg.GetClustersByUser(email)
		if err != nil {
			log.Println(email, "don't belong to any cluster")
		}
		for i := range clusters {
			clusters[i].Token = "******"
		}
		log.Println(email, "list the clusters")
	} else {
		clusters, _ = pkg.GetAllClusters()
		log.Println(email, "list the clusters as an admin")
	}

	res = pkg.ClusterResponse{
		Status:   "success",
		Message:  "Cluster By user",
		Clusters: clusters,
	}
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandleDeleteCluster(w http.ResponseWriter, r *http.Request) {
	userType, mail, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(mail, "try to delete a cluster but he's not an admin")
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
	name := r.FormValue("name")

	cluster := pkg.ClusterExists(name)
	if !cluster {
		res = pkg.Response{
			Status:  "fail",
			Message: "Cluster don't exist!",
		}
	} else {
		if pkg.CountDirectories("clusters/"+name) > 0 {
			res = pkg.Response{
				Status:  "fail",
				Message: "You can't delete this cluster, it contains groups or/and users!",
			}
			log.Println(mail, "try to delete a cluster that have users or/and groups")
		} else {
			err = os.RemoveAll("clusters/" + name)
			if err != nil {
				log.Println(err)
			}
			_ = pkg.DeleteCluster(name)
			res = pkg.Response{
				Status:  "success",
				Message: "The cluster " + name + " is removed successfully",
			}
			log.Println(mail, "delete the cluster", name)
		}

	}
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}
