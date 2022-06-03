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

package config

import (
	"strconv"

	"github.com/thatpix3l/fntwo/obj"
)

// Config used during the start of the application
type App struct {
	VmcListenIP          string `json:"vmc_listen_ip"`          // IP address to listen for VMC data
	VmcListenPort        int    `json:"vmc_listen_port"`        // Port to listen for VMC data
	WebServeIP           string `json:"web_listen_ip"`          // IP address to serve the frontend
	WebServePort         int    `json:"web_listen_port"`        // Port to serve the frontend
	ModelUpdateFrequency int    `json:"model_update_frequency"` // Times per second model transformation is sent to clients
	SceneCfgFile         string `json:"scene_config_file"`      // Path to scene config file
	AppCfgFile           string `json:"app_config_file"`        // Path to app config file
	VRMFile              string `json:"vrm_file"`               // Path to VRM that will be loaded and overwritten
}

// Config used for the looks and appearance of the model viewer.
// This is what most people will care about.
type Scene struct {
	Camera obj.Camera `json:"camera"`
}

// Get the combined string of VMCListenIP and VMCListenPort
func (appCfg App) GetVmcServerAddress() string {
	return appCfg.VmcListenIP + ":" + strconv.Itoa(appCfg.VmcListenPort)
}

// Get the combined string of WebServeIP and WebServePort
func (appCfg App) GetWebServerAddress() string {
	return appCfg.WebServeIP + ":" + strconv.Itoa(appCfg.WebServePort)
}
