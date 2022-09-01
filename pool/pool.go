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

type updateCallback func(client *Client)

type Pool struct {
	clients map[string]Client // List of clients waiting for data
}

type Client struct {
	ID       string         // Unique client ID
	callback updateCallback // Callback to run on new relayedData
	poolPtr  *Pool          // Pointer to an existing pool
}

// Create pool client, with callback that gets run everytime pool's Update is called
func (p Pool) Create(process updateCallback) {

	clientID := helper.RandomString(8)
	newClient := Client{
		ID:       clientID,
		callback: process,
		poolPtr:  &p,
	}
	p.clients[clientID] = newClient

}

// Delete client
func (c Client) Delete() {
	delete(c.poolPtr.clients, c.ID)
}

// Log the amount of clients in pool
func (p Pool) LogCount() {
	log.Printf("Number of clients: %d", len(p.clients))
}

// Run each client's update callback
func (p Pool) Update() {
	for _, c := range p.clients {
		c.callback(&c)
	}
}

// Return a new pool manager
func New() Pool {
	return Pool{
		clients: make(map[string]Client),
	}
}
