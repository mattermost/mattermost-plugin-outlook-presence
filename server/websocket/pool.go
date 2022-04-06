package websocket

import (
	"fmt"

	"github.com/Brightscout/mattermost-plugin-outlook-presence/server/serializer"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan *serializer.StatusChangedEvent
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *serializer.StatusChangedEvent),
	}
}

func (p *Pool) Start(api plugin.API) {
	for {
		select {
		case client := <-p.Register:
			p.Clients[client] = true
			api.LogInfo(fmt.Sprintf("Size of connection pool: %d", len(p.Clients)))
		case client := <-p.Unregister:
			delete(p.Clients, client)
			api.LogInfo(fmt.Sprintf("Size of connection pool: %d", len(p.Clients)))
		case statusChangedEvent := <-p.Broadcast:
			api.LogInfo("Sending message to all clients in pool")
			for client, _ := range p.Clients {
				if err := client.Conn.WriteJSON(statusChangedEvent); err != nil {
					api.LogError("error in broadcasting the status changed event.", "Error", err.Error())
				}
			}
		}
	}
}