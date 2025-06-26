// handles Websocket connections and calls Gemini
package main

import (
	"context"
	"log"
	"net/http"
	"fmt"
	"encoding/json"

	"github.com/gorilla/websocket"
)

// TODO: use environment variables for this instead
const ProjectId = "allmind-carta-proj-1"
const InstanceName = "allmind-f25-take-home-db"
const TableName = "chat-messages"

// TODO: [answer] is *name a pointer?
// TODO [answer] what is the return type of websocket.Upgrader?
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow any origin for dev
	},
}


func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("handle websocket running")
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Websocket upgrade error:", err)
		return
	}
	// TODO: understand what defer keyword means
	defer wsConn.Close()

	ctx := context.Background()
	session, err := NewGeminiChatSession(ctx) // will define NewGeminiChatSession in another file
	if err != nil {
		log.Fatal("Gemini init error:", err) // aah Println can handle multipe arguments. Good times.
		return
	}

	// connect to Google Big table using NewBigTableWriter from another file
	bigtable, err := NewBigTableWriter(ctx, ProjectId, InstanceName, TableName)

	for  {
		// receiving a message from user
		_, msg, err := wsConn.ReadMessage() // TODO: is this something given by Gorilla?
		if err != nil {
			log.Println("Read error", err)
			break
		}
		log.Println("Received msg from frontend: ", string(msg))
		
		err = bigtable.WriteMessage(ctx, "user", string(msg))
		var tempMsg string
		if err != nil {
			tempMsg = fmt.Sprintf("error while writing user's message to bigtable %s", err.Error())
		} else {
			tempMsg = "[Big Table] Wrote user's msg"
		}
		log.Println(tempMsg)
		
		resp, err := session.SendMessage(ctx, string(msg)) // sending it to Gemini
		if err != nil {
			log.Println("Gemini error:", err)
			wsConn.WriteMessage(websocket.TextMessage, []byte("Error from Gemini"))
			continue
		}
		
		// TODO [answer] figure out whether this is type casting
		wsConn.WriteMessage(websocket.TextMessage, [](byte)(resp)) 
		err = bigtable.WriteMessage(ctx, "ai", resp)
		
		if err != nil {
			tempMsg = fmt.Sprintf("error while writing AI's message to bigtable", err)
		} else {
			tempMsg = "[Big Table] Wrote AI's msg"
		}
		log.Println(tempMsg)
	}
}


func handleHistory(w http.ResponseWriter, r *http.Request) {
	log.Println("[handleHistory]: called")

	ctx := context.Background()
	
	// connect to Google Big table using NewBigTableWriter from another file
	bigtable, err := NewBigTableWriter(ctx, ProjectId, InstanceName, TableName)
	if err != nil {
		http.Error(w, "Bigtable error in /history when opening custom object", 500)
		return
	}
	
	messages, err := bigtable.ReadMessages(ctx)
	if err != nil {
		http.Error(w, "Failed to read message", 500)
		log.Fatal("Failed when reading message: ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
