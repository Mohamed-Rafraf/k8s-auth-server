package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mohamed-rafraf/k8s-auth-server/pkg"
)

func HandleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		HandlePostGroup(w, r)
	case http.MethodDelete:
		HandleDeleteGroup(w, r)
	case http.MethodGet:
		HandleGetGroup(w, r)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"status": "error", "message": "Method not allowed"}`))
	}
}

func HandlePostGroup(w http.ResponseWriter, r *http.Request) {
	userType, mail, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(mail, "is not authorized to create group")
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
	cluster := r.FormValue("cluster")
	exist := pkg.DirectoryExists("clusters/" + cluster + "/" + name)

	if exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "Group is already Exists!",
		}
	} else {

		err = os.Mkdir("clusters/"+cluster+"/"+name, 0755)
		if err != nil {
			panic(err)
		}
		// Get the uploaded file from the form data
		err = pkg.UploadFile("clusters/"+cluster+"/"+name+"/RBAC.yaml", r)
		if err != nil {
			panic(err)
		}

		fmt.Println(mail, "create new group called", name, "in", cluster)

		res = pkg.Response{
			Status:  "success",
			Message: "The Group " + name + " is created successfully in " + cluster,
		}

	}
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	userType, mail, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(mail, "is not authorized to delete a group")
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
	cluster := r.FormValue("cluster")

	exist := pkg.DirectoryExists("clusters/" + cluster + "/" + name)
	if !exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "Group Don't Exist",
		}
		log.Println(mail, "try to delete a group that don't exist")
	} else {
		if pkg.CountDirectories("clusters/"+cluster+"/"+name) > 0 {
			res = pkg.Response{
				Status:  "fail",
				Message: "You can't delete this group, it contains users!",
			}
			log.Println(mail, "try to delete a group that have users")
		} else {
			err := os.RemoveAll("clusters/" + cluster + "/" + name)
			if err != nil {
				log.Println(err)
			}
			res = pkg.Response{
				Status:  "success",
				Message: "The group " + name + " in cluster " + cluster + " is removed successfully",
			}
			log.Println(mail, "delete the group", name, "inside", cluster)

		}

	}
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	userType, mail, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	var res pkg.Response
	encoder := json.NewEncoder(w)
	if userType != "admin" {
		log.Println(mail, "is not authorized to list the groups")
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

	name := r.FormValue("cluster")
	cluster := pkg.ClusterExists(name)

	if !cluster {
		res = pkg.Response{
			Status:  "fail",
			Message: "Cluster don't exist",
		}
	} else {
		group_names, _ := pkg.GetSubdirs("clusters/" + name)
		var groups []string
		for _, group := range group_names {
			if !strings.Contains(group, "@") {
				parts := strings.Split(group, "/")
				groups = append(groups, parts[len(parts)-1]+"-"+strconv.Itoa(pkg.CountDirectories("clusters/"+name+"/"+parts[len(parts)-1])))
			}

		}
		res = pkg.Response{
			Status:  "success",
			Message: strings.Join(groups, ", "),
		}

	}

	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}
}
