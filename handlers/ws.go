package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mohamed-rafraf/k8s-auth-server/pkg"
)

type Connection struct {
	ws   *websocket.Conn
	name string
}

func (c *Connection) Write(p []byte) (n int, err error) {
	err = c.ws.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

var Connections = make(map[string]*Connection)

func HandleIncomingMessages(conn *Connection) {
	for {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			delete(Connections, conn.name)
			break
		}
		log.Printf("Received message from %s: %s\n", conn.name, string(message))
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	clusterName := r.URL.Query().Get("clusterName")
	token := r.URL.Query().Get("token")
	cluster, err := pkg.GetClusterByName(clusterName)
	log.Println("Trying to open a web socket")
	if token == "" {
		log.Println("Token not found", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Println(err.Error(), http.StatusBadRequest)
		return
	}
	if &cluster == nil {
		log.Println("This cluster don't exist", http.StatusBadRequest)
		return
	}
	if cluster.Token != token {
		log.Println("Invalid cluster TOKEN !!", http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	conn := &Connection{ws: ws, name: clusterName}
	Connections[clusterName] = conn

	log.Println("WebSocket Connection established with cluster:", clusterName)

	// Handle incoming messages from the client
	//go HandleIncomingMessages(conn)

	// Send a welcome message to the client
	welcomeMessage := fmt.Sprintf("Welcome to the server, %s!", clusterName)
	err = conn.ws.WriteMessage(websocket.TextMessage, []byte(welcomeMessage))
	if err != nil {
		log.Println("WebSocket write error:", err)
	}
}
func SendMsg(w http.ResponseWriter, r *http.Request) {
	clusterName := r.FormValue("clusterName")
	msg := r.FormValue("message")
	_, err := SendMessageToClient(clusterName, msg)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Message sent to", clusterName)

}

func SendMessageToClient(clusterName string, message string) (string, error) {
	conn, ok := Connections[clusterName]
	if !ok {
		return "", fmt.Errorf("Connection not found for cluster %s", clusterName)
	}
	err := conn.ws.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return "", err
	}
READ:
	_, msg, err := conn.ws.ReadMessage()
	if err != nil {
		log.Println(err)
		goto READ
	}
	if string(msg) == "" {
		goto READ
	}
	return string(msg), nil

}

func HandleMsg(w http.ResponseWriter, h *http.Request) {
	switch h.Method {
	//case http.MethodGet:
	case http.MethodPost:
		SendMsg(w, h)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Invalid request method")
	}
}
