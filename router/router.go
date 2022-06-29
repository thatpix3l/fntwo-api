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

package router

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/frontend"
	"github.com/thatpix3l/fntwo/obj"
	"github.com/thatpix3l/fntwo/pool"
)

var (
	appCfg   *config.App
	sceneCfg *config.Scene
	liveVRM  *obj.VRM
)

// Helper function to upgrade an HTTP connection to WebSockets
func wsUpgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return ws, err
}

// Helper func to allow all origin, headers, and methods for HTTP requests.
func allowHTTPAllPerms(wPtr *http.ResponseWriter) {

	// Set CORS policy
	(*wPtr).Header().Set("Access-Control-Allow-Origin", "*")
	(*wPtr).Header().Set("Access-Control-Allow-Methods", "*")
	(*wPtr).Header().Set("Access-Control-Allow-Headers", "*")

}

// Save a given scene to the default path
func saveScene(scene *config.Scene) error {

	// Convert the scene config in memory into bytes
	sceneCfgBytes, err := json.MarshalIndent(scene, "", " ")
	if err != nil {
		return err
	}

	// Store config bytes into file
	if err := os.WriteFile(appCfg.SceneFilePath, sceneCfgBytes, 0644); err != nil {
		return err
	}

	return nil

}

func New(appConfigPtr *config.App, sceneCfgPtr *config.Scene, vrmPtr *obj.VRM) *mux.Router {

	appCfg = appConfigPtr
	sceneCfg = sceneCfgPtr
	liveVRM = vrmPtr

	router := mux.NewRouter()

	// Create new camera pool, listen in background for messages
	cameraPool := pool.New()
	go cameraPool.Listen()

	// Route for relaying the internal state of the camera to all clients
	router.HandleFunc("/client/camera", func(w http.ResponseWriter, r *http.Request) {

		// Upgrade GET request to WebSocket
		ws, err := wsUpgrade(w, r)
		if err != nil {
			log.Println(err)
		}

		// On first-time WebSocket connection, update all existing clients
		cameraPool.Update(sceneCfg.Camera)

		// Add a new pool client
		log.Println("Adding new client...")
		cameraClient := cameraPool.Create(func(relayedData interface{}) {

			var ok bool // Boolean for if the pool's relayedData was type asserted as obj.Camera
			if sceneCfg.Camera, ok = relayedData.(obj.Camera); !ok {
				log.Fatal("Severe error from type asserting camera pool data as obj.Camera! Exiting...")
			}

			// Write camera data to connected frontend client
			if err := ws.WriteJSON(sceneCfg.Camera); err != nil {
				log.Println(err)
			}

		})

		cameraPool.LogCount() // Log count of clients before reading data

		for {

			// Wait for and read new camera data from WebSocket client
			if err := ws.ReadJSON(&sceneCfg.Camera); err != nil {
				log.Printf("Error reading WebSocket for client with ID %s, removing dead client...", cameraClient.ID)
				cameraClient.Delete()
				break
			}

			// Update camera pool with new camera data
			log.Println("Updating camera data for camera pool...")
			cameraPool.Update(sceneCfg.Camera)

		}

		cameraPool.LogCount() // Log count of clients after reading data

	})

	// Route for updating VRM model data to all clients
	router.HandleFunc("/client/model", func(w http.ResponseWriter, r *http.Request) {

		ws, err := wsUpgrade(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		// Forever send to client the VRM data
		for {

			if err := ws.WriteJSON(*liveVRM); err != nil {
				return
			}
			time.Sleep(time.Duration(1e9 / appCfg.ModelUpdateFrequency))

		}

	})

	// Route for getting the default VRM model
	router.HandleFunc("/api/model", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to retrieve default VRM file")

		// Set model name and CORS policy
		w.Header().Set("Content-Disposition", "attachment; filename=default.vrm")
		allowHTTPAllPerms(&w)

		// Serve default VRM file
		http.ServeFile(w, r, appCfg.VRMFilePath)

	}).Methods("GET", "OPTIONS")

	// Route for setting the default VRM model
	router.HandleFunc("/api/model", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to set default VRM file")

		allowHTTPAllPerms(&w)

		// Destination VRM file on system
		dest, err := os.Create(appCfg.VRMFilePath)
		if err != nil {
			log.Println(err)
			return
		}

		// Copy request body binary to destination on system
		if _, err := io.Copy(dest, r.Body); err != nil {
			log.Println(err)
			return
		}

	}).Methods("PUT", "OPTIONS")

	// HTTP PUT request route for saving the internal state of the scene config
	router.HandleFunc("/api/config/scene", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to save current scene")

		// Access control
		allowHTTPAllPerms(&w)

		if err := saveScene(sceneCfg); err != nil {
			log.Println(err)
			return
		}

	}).Methods("PUT", "OPTIONS")

	// HTTP route for retrieving the initial config for the server
	router.HandleFunc("/api/config/app", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to retrieve initial config")

		// Access control
		allowHTTPAllPerms(&w)

		// Marshal initial config into bytes
		initCfgBytes, err := json.Marshal(appCfg)
		if err != nil {
			log.Println(err)
			return
		}

		// Reply back to request with byte-format of initial config
		w.Header().Set("Content-Type", "application/json")
		w.Write(initCfgBytes)

	}).Methods("GET", "OPTIONS")

	// All other requests are sent to the embedded web frontend
	frontendRoot, err := frontend.FS()
	if err != nil {
		log.Fatal(err)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(frontendRoot)))

	return router

}
