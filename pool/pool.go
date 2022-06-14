package pool

import (
	"log"

	"github.com/thatpix3l/fntwo/helper"
)

type Pool struct {
	clients       map[string]Client // List of clients waiting for data
	globalChannel chan interface{}  // Global channel that any client can relay data through
}

type Client struct {
	ID      string
	channel chan interface{}
	poolPtr *Pool
}

// Create client, return reference to it
func (p Pool) Create() *Client {

	clientID := helper.RandomString(8)
	newClient := Client{
		ID:      clientID,
		channel: make(chan interface{}),
		poolPtr: &p,
	}
	p.clients[clientID] = newClient

	return &newClient

}

// Delete client
func (c Client) Delete() {
	delete(c.poolPtr.clients, c.ID)
}

// Blocking read data from client
func (c Client) Read() interface{} {
	return <-c.channel
}

// Log the amount of clients in pool
func (p Pool) LogCount() {
	log.Printf("Number of clients: %d", len(p.clients))
}

// Update pool of clients with the given data
func (p Pool) Update(data interface{}) {
	p.globalChannel <- data
}

// Blocking loop read data from pool and relay it to all clients
func (p Pool) Listen() {
	for {

		// Wait for and read data from some client
		data := <-p.globalChannel

		// Relay data to all clients
		for _, c := range p.clients {
			c.channel <- data
		}

	}
}

// Return a new pool manager
func New() Pool {
	return Pool{
		clients:       make(map[string]Client),
		globalChannel: make(chan interface{}),
	}
}
