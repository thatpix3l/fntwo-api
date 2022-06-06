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
	VmcListenIP          string `json:"vmc_listen_ip"`          // Address interface the VMC server listens on
	VmcListenPort        int    `json:"vmc_listen_port"`        // Port the VMC server listens on
	WebListenIP          string `json:"web_listen_ip"`          // Address interface the frontend and API listens on
	WebListenPort        int    `json:"web_listen_port"`        // Port the frontend and API listens on
	ModelUpdateFrequency int    `json:"model_update_frequency"` // Times per second model transformation is sent to clients
	SceneDirPath         string `json:"scene_home"`             // Path to scene directory
	SceneFilePath        string `json:"scene_config_file"`      // Path to scene config file
	AppCfgFilePath       string `json:"app_config_file"`        // Path to app config file
	VRMFilePath          string `json:"vrm_file"`               // Path to VRM that will be loaded and overwritten
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
	return appCfg.WebListenIP + ":" + strconv.Itoa(appCfg.WebListenPort)
}

type MotionReceiver struct {
	VRM    *obj.VRM // Pointer to an existing VRM struct to apply transformation towards
	AppCfg *App     // Pointer to an existing app config, for reading various settings
	Start  func()   // Generic function to start a given tracker type. Up to implementation on what's actually called
}
