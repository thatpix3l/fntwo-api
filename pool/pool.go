/*
fntwo: An easy to use tool for VTubing
Copyright (C) 2022 thatpix3l <contact@thatpix3l.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, version 3 of the License.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
	for _, c := range p.clients {
		c.process(data)
	}
}

// Return a new pool manager
func New() Pool {
	return Pool{
		clients:       make(map[string]Client),
		globalChannel: make(chan interface{}),
	}
}
