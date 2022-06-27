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
	ID      string                        // Unique client ID
	process func(relayedData interface{}) // Function to process relayed data from pool's global channel
	poolPtr *Pool                         // Pointer to an existing pool
}

// Create client, return reference to it
func (p Pool) Create(dataCallback func(relayedData interface{})) *Client {

	clientID := helper.RandomString(8)
	newClient := Client{
		ID:      clientID,
		process: dataCallback,
		poolPtr: &p,
	}
	p.clients[clientID] = newClient

	return &newClient

}

// Delete client
func (c Client) Delete() {
	delete(c.poolPtr.clients, c.ID)
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

		// Relay data to each client's individual function
		for _, c := range p.clients {
			c.process(data)
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
