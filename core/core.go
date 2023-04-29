package core

import (
	"crypto/tls"
	"log"
	"net/http"

	//"github.com/mohamed-rafraf/k8s-auth-server/pkg"
	"github.com/mohamed-rafraf/k8s-auth-server/handlers"
)

/*func main() {
	http.HandleFunc("/admin",func (w http.ResponseWriter, r *http.Request) {handlers.HandleGoogleLogin(w,r,"admin")})
	http.HandleFunc("/login",func (w http.ResponseWriter, r *http.Request) {handlers.HandleGoogleLogin(w,r,"login")})
	http.HandleFunc("/users", handlers.HandleUsers)
	http.HandleFunc("/groups", handlers.HandleGroups)
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	http.HandleFunc("/verify", handlers.HandleVerify)
	http.HandleFunc("/auth", handlers.HandleAuth)
	http.HandleFunc("/msg",handlers.HandleMsg)
	http.HandleFunc("/clusters", handlers.HandleClusters)
	http.HandleFunc("/callback", handlers.HandleGoogleCallback)
	http.Handle("/test", OAuth2MiddlewareforLogin(http.HandlerFunc(handlers.HandleTest)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}*/

func StartServer() error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) { handlers.HandleGoogleLogin(w, r, "admin") })
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { handlers.HandleGoogleLogin(w, r, "login") })
	http.HandleFunc("/users", handlers.HandleUsers)
	http.HandleFunc("/groups", handlers.HandleGroups)
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	http.HandleFunc("/verify", handlers.HandleVerify)
	http.HandleFunc("/auth", handlers.HandleAuth)
	http.HandleFunc("/msg", handlers.HandleMsg)
	http.HandleFunc("/clusters", handlers.HandleClusters)
	http.HandleFunc("/callback", handlers.HandleGoogleCallback)
	http.HandleFunc("/permissions", handlers.HandlePermissions)
	log.Fatal(http.ListenAndServe(":8080", nil))
	return nil
}
