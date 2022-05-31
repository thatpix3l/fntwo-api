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

import "strconv"

type Keys struct {
	VmcListenIP          string // IP address to listen for VMC data
	VmcListenPort        int    // Port to listen for VMC data
	WebServeIP           string // IP address to serve the frontend
	WebServePort         int    // Port to serve the frontend
	ModelUpdateFrequency int    // Times per second to send the live model data to frontend clients
	ConfigPath           string // Path to config file
}

func (k Keys) GetVmcSocketAddress() string {
	return k.VmcListenIP + ":" + strconv.Itoa(k.VmcListenPort)
}

func (k Keys) GetWebSocketAddress() string {
	return k.WebServeIP + ":" + strconv.Itoa(k.WebServePort)
}
