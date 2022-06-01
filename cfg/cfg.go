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

package cfg

import (
	"strconv"

	"github.com/thatpix3l/fntwo/obj"
)

// Config used during the start of the program
type Initial struct {
	VmcListenIP          string // IP address to listen for VMC data
	VmcListenPort        int    // Port to listen for VMC data
	WebServeIP           string // IP address to serve the frontend
	WebServePort         int    // Port to serve the frontend
	ModelUpdateFrequency int    // Times per second to send the live model data to frontend clients
	RuntimeCfgPath       string // Path to runtime config file
	ConfigPath           string // Path to initial config file
}

// Config used during the runtime of the program
type Runtime struct {
	Camera obj.Camera `json:"camera"`
}

// Get the combined string of VMCListenIP and VMCListenPort
func (initCfg Initial) GetVmcSocketAddress() string {
	return initCfg.VmcListenIP + ":" + strconv.Itoa(initCfg.VmcListenPort)
}

// Get the combined string of WebServeIP and WebServePort
func (initCfg Initial) GetWebSocketAddress() string {
	return initCfg.WebServeIP + ":" + strconv.Itoa(initCfg.WebServePort)
}
