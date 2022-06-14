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

package app

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/obj"
	"github.com/thatpix3l/fntwo/receivers/facemotion3d"
	"github.com/thatpix3l/fntwo/receivers/virtualmotioncapture"
	"github.com/thatpix3l/fntwo/router"
)

var (
	liveVRM = &obj.VRM{
		Bones:       obj.HumanBodyBones{},
		BlendShapes: obj.BlendShapes{},
	} // VRM transformation data, updated from sources
	appCfg   *config.App       // Initial config for settings of the app
	sceneCfg = &config.Scene{} // Scene config for various live data
)

// Helper func to load a scene config file
func loadScene(sceneCfgPath string) error {

	// Read in a scene JSON config file
	content, err := os.ReadFile(sceneCfgPath)
	if err != nil {
		return err
	}

	// Unmarshal into memory's scene config
	if err := json.Unmarshal(content, sceneCfg); err != nil {
		return err
	}

	return nil

}

// Attempt to create a new, default scene if no scene already exists
func saveDefaultScene() error {

	// Open and create scene file
	sceneFile, err := os.OpenFile(appCfg.SceneFilePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer sceneFile.Close()

	// Create the scene that will be used as the default
	defaultScene := config.Scene{
		Camera: obj.Camera{
			GazeTowards: obj.Position{
				X: 0,
				Y: 0,
				Z: 0,
			},
			GazeFrom: obj.Position{
				X: 3,
				Y: 3,
				Z: 3,
			},
		},
	}

	// Marshal defaultScene into bytes
	defaultSceneBytes, err := json.MarshalIndent(defaultScene, "", "    ")
	if err != nil {
		return err
	}

	// Write to file the scene data
	if _, err := sceneFile.Write(defaultSceneBytes); err != nil {
		return err
	}

	return nil

}

// Entrypoint
func Start(initialConfig *config.App) {

	// Store pointer of generated config file to use throughout this program
	appCfg = initialConfig

	// If needed, create a default scene file
	if err := saveDefaultScene(); err != nil {
		log.Println(err)
	}

	// Load scene config from disk
	if err := loadScene(appCfg.SceneFilePath); err != nil {
		log.Println(err)
	}

	// Background listen and serve for face and bone tracking
	vmcServer := virtualmotioncapture.New(liveVRM, appCfg)
	fm3dServer := facemotion3d.New(liveVRM, appCfg)
	go vmcServer.Start()
	go fm3dServer.Start()

	// Blocking listen and serve for WebSockets and API server
	log.Printf("Serving frontend and listening for clients/API queries on %s", appCfg.GetWebServerAddress())
	routerAPI := router.New(appCfg, sceneCfg, liveVRM)
	http.ListenAndServe(appCfg.GetWebServerAddress(), routerAPI)

}
