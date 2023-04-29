package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	Url "net/url"
	"os"
	"strings"

	"github.com/mohamed-rafraf/k8s-auth-server/pkg"
	"golang.org/x/oauth2"
)

func HandleAuth(w http.ResponseWriter, r *http.Request) {

	var data, rbac, rbacFile string
	var res pkg.Response
	var content []byte

	userType, mail, msg, err := isAuthorized(w, r)
	if err != nil {
		log.Println(err)
		return
	}

	encoder := json.NewEncoder(w)
	if userType != "user" {
		log.Println(mail, "is not authorized to authenticate clusters")
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

	cluster := r.FormValue("cluster")
	cluster_path := "clusters/" + cluster + "/"
	group, _ := pkg.GetGroupFromMail(cluster_path, mail)
	cluster_exist := pkg.ClusterExists(cluster)

	if cluster == "" {
		res = pkg.Response{
			Status:  "fail",
			Message: "Please specify the cluster",
		}
		log.Println(mail, "Try to authenticate without specify the cluster name")
		goto SEND
	}

	if !cluster_exist {
		res = pkg.Response{
			Status:  "fail",
			Message: "The cluster is not exist!",
		}
		log.Println(mail, "try to authenticate a cluster that didn't exist")
		goto SEND
	}

	if group == "" {
		rbacFile = cluster_path + mail + "/RBAC.yaml"
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
	rbac = string(content)

	data, err = SendMessageToClient(cluster, "Token-Create,user="+mail+",rbac="+base64.StdEncoding.EncodeToString([]byte(rbac)))

	if err != nil {
		res = pkg.Response{
			Status:  "fail",
			Message: "Failed to send request",
		}
		log.Println(err)
		goto SEND
	}

	res = pkg.Response{
		Status:  "success",
		Message: data,
	}
SEND:
	err = encoder.Encode(res)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandleVerify(w http.ResponseWriter, r *http.Request) {

	var response pkg.Response
	encoder := json.NewEncoder(w)
	name := r.FormValue("cluster")
	token := r.FormValue("token")
	api := r.FormValue("api")
	cluster, err := pkg.GetClusterByName(name)
	if err != nil {
		response = pkg.Response{
			Status:  "fail",
			Message: "Can't Check the cluster",
		}
		log.Println(err)
		goto SEND
	}
	if &cluster == nil {
		response = pkg.Response{
			Status:  "fail",
			Message: "Cluster didn't exist",
		}
		goto SEND
	}

	if cluster.Token != token {
		response = pkg.Response{
			Status:  "fail",
			Message: "Your token is wrong!",
		}
		goto SEND
	}

	if cluster.Status {
		response = pkg.Response{
			Status:  "success",
			Message: "The Cluster is already activated",
		}
		goto SEND

	}

	err = pkg.UpdateCluster(name, api)
	if err != nil {
		response = pkg.Response{
			Status:  "fail",
			Message: "Can't activate the cluster",
		}
		log.Println(err)
		goto SEND
	}

	response = pkg.Response{
		Status:  "success",
		Message: "The cluster is activated successfuly",
	}

SEND:
	err = encoder.Encode(response)
	if err != nil {
		log.Println(err)
		return
	}

}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request, state string) {

	url := pkg.GoogleOauthConfig.AuthCodeURL(state)
	response := pkg.Response{
		Status:  "success",
		Message: base64.StdEncoding.EncodeToString([]byte(url)),
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		log.Println(err)
	}
	log.Println("Someone try to login")
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	content, err := Authenticate(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		return
	}
	fmt.Fprintf(w, "Content: %s\n", content)
}

func Authenticate(state string, code string) (string, error) {
	if state != "login" && state != "admin" {
		return "nil", fmt.Errorf("invalid oauth state")
	}

	token, err := pkg.GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return "nil", fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return "nil", fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	var user pkg.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		panic(err)
	}

	if err != nil {
		return "nil", fmt.Errorf("failed reading response body: %s", err.Error())
	}
	if state == "admin" {
		admin := pkg.AdminExists(user.Email)
		if !admin {
			HandleLogout(token.AccessToken)
			return "YOUR NOT AUTHORIZED, ARE YOU ADMIN BRO ??? ", nil
		}
	} else {
		user := pkg.UserExists(user.Email)
		if !user {
			HandleLogout(token.AccessToken)
			return "YOUR NOT AUTHORIZED, ARE YOU REGISTRED BRO ??? ", nil
		}
	}

	return "YOU NEED THIS ONE! COPY IT! " + token.AccessToken, nil
}

func revokeToken(token *oauth2.Token) error {
	oauthClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	resp, err := oauthClient.PostForm("https://accounts.google.com/o/oauth2/revoke", Url.Values{
		"token": []string{token.AccessToken},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to revoke token, status code: %d", resp.StatusCode)
	}

	return nil
}

func HandleLogout(accessToken string) error {
	token := &oauth2.Token{AccessToken: accessToken}
	if err := revokeToken(token); err != nil {
		return err
	}
	return nil
}

func isAuthorized(w http.ResponseWriter, r *http.Request) (string, string, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "none", "none", "Missing authorization header", nil
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	if accessToken == "" {
		return "none", "none", "Missing access token", nil
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return "none", "none", "Invalid Token", err
	}

	if response.Status != "200 OK" {
		return "none", "none", "Unauthorized Token", nil
	}

	defer response.Body.Close()

	data, _ := io.ReadAll(response.Body)
	var user pkg.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		return "none", "none", "JSON DADY", err
	}

	admin := pkg.AdminExists(user.Email)
	if !admin {

		User := pkg.UserExists(user.Email)
		if !User {
			HandleLogout(accessToken)
			return "none", "none", "YOUR NOT AUTHORIZED, ARE YOU REGISTRED BRO ??? ", nil
		} else {
			return "user", user.Email, "You are an authorized user", nil
		}
	} else {
		return "admin", user.Email, "YOU ARE ADMIN", nil
	}
}
