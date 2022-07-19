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
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/thatpix3l/fntwo/obj"
)

type Address string

// Return primitive string type of Address
func (a *Address) String() string {
	return string(*a)
}

// Return just the IP of the full address
func (a *Address) IP() string {
	fullAddress := strings.Split(string(*a), ":")
	return fullAddress[0]
}

// Return just the Port of the full address
func (a *Address) Port() int {

	fullAddress := strings.Split(string(*a), ":")
	port, err := strconv.Atoi(fullAddress[1])
	if err != nil {
		log.Fatal(err)
	}

	return port

}

// Setter, mainly used for cobra
func (a *Address) Set(v string) error {
	*a = Address(v)
	return nil
}

// Retrieve type, mainly used for cobra
func (a *Address) Type() string {
	return "address"
}

// Config used during the start of the application
type App struct {
	VMCListen            Address `json:"vmc_listen"`             // Address interface the VMC server listens on
	FM3DListen           Address `json:"fm3d_listen"`            // Address interface the Facemotion3D server listens on
	FM3DDevice           Address `json:"fm3d_device"`            // IP address of phone/device to tell to start sending Facemotion3D data
	APIListen            Address `json:"api_listen"`             // Address interface the API server listens on
	ModelUpdateFrequency int     `json:"model_update_frequency"` // Times per second the model transformation data is sent to clients
	SceneDirPath         string  `json:"scene_home"`             // Path to scene directory
	SceneFilePath        string  `json:"scene_file"`             // Path to scene config file
	AppCfgFilePath       string  `json:"config_file"`            // Path to app config file
	VRMFilePath          string  `json:"vrm_file"`               // Path to VRM file
}

// Return the tag of a field, with underscores replaced with dashes
func (a App) TagWithDashes(name string) string {
	structField, _ := reflect.TypeOf(a).FieldByName(name)
	tagData := strings.ReplaceAll(structField.Tag.Get("json"), "_", "-")
	return tagData
}

// Config used for the looks and appearance of the model viewer.
// This is what most people will care about.
type Scene struct {
	Camera obj.Camera `json:"camera"`
}
