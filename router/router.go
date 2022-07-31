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
	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/frontend"
	"github.com/thatpix3l/fntwo/helper"
	"github.com/thatpix3l/fntwo/obj"
	"github.com/thatpix3l/fntwo/pool"
	"github.com/thatpix3l/fntwo/receivers"
)

var (
	activeReceiver *receivers.MotionReceiver
)

// Helper func to allow all origin, headers, and methods for HTTP requests.
func allowHTTPAllPerms(wPtr *http.ResponseWriter) {

	// Set CORS policy
	(*wPtr).Header().Set("Access-Control-Allow-Origin", "*")
	(*wPtr).Header().Set("Access-Control-Allow-Methods", "*")
	(*wPtr).Header().Set("Access-Control-Allow-Headers", "*")

}

// Save a given scene to the default path
func saveScene(scene *config.Scene, sceneFilePath string) error {

	// Convert the scene config in memory into bytes
	sceneCfgBytes, err := json.MarshalIndent(scene, "", " ")
	if err != nil {
		return err
	}

	// Store config bytes into file
	if err := os.WriteFile(sceneFilePath, sceneCfgBytes, 0644); err != nil {
		return err
	}

	return nil

}

func New(appCfg *config.App, sceneCfg *config.Scene, receiverMap map[string]*receivers.MotionReceiver) *mux.Router {

	// Use the first motion receiver in map
	for name, receiver := range receiverMap {
		log.Printf("The active receiver is %s", name)
		activeReceiver = receiver
		break
	}

	// Router for API and web frontend
	router := mux.NewRouter()

	// Route for relaying the internal state of the camera to all clients
	cameraPool := pool.New()
	router.HandleFunc("/live/read/camera", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Adding new camera client...")

		// Upgrade GET request to WebSocket
		ws, err := helper.WebSocketUpgrade(w, r)
		if err != nil {
			log.Println(err)
		}

		// Add a new camera client
		cameraPool.Create(func(relayedData interface{}, client *pool.Client) {

			var ok bool // Boolean for if the pool's relayedData was type asserted as obj.Camera
			if sceneCfg.Camera, ok = relayedData.(obj.Camera); !ok {
				log.Println("Couldn't type assert relayed data as a camera")
				return
			}

			// Write camera data to connected frontend client
			if err := ws.WriteJSON(sceneCfg.Camera); err != nil {
				log.Println(err)
				client.Delete()
				ws.Close()
			}

		})

	})

	router.HandleFunc("/live/write/camera", func(w http.ResponseWriter, r *http.Request) {

		ws, err := helper.WebSocketUpgrade(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		for {

			if err := ws.ReadJSON(&sceneCfg.Camera); err != nil {
				return
			}

			cameraPool.Update(sceneCfg.Camera)

		}

	})

	// Route for updating VRM model data to all clients
	router.HandleFunc("/live/read/model", func(w http.ResponseWriter, r *http.Request) {

		// Upgrade model data client into a WebSocket
		ws, err := helper.WebSocketUpgrade(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		for {

			// Process and send the VRM data to WebSocket
			activeReceiver.VRM.Read(func(vrm *obj.VRM) {

				// Send VRM data to WebSocket client
				if err := ws.WriteJSON(*vrm); err != nil {
					return
				}

			})

			// Wait for whatever how long, per second. By default, 1/60 of a second
			time.Sleep(time.Duration(1e9 / appCfg.ModelUpdateFrequency))

		}

	})

	// Route for getting the default VRM model
	router.HandleFunc("/api/read/model/get", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to retrieve default VRM file")

		// Set model name and CORS policy
		w.Header().Set("Content-Disposition", "attachment; filename=default.vrm")
		allowHTTPAllPerms(&w)

		// Serve default VRM file
		http.ServeFile(w, r, appCfg.VRMFilePath)

	}).Methods("GET", "OPTIONS")

	// Route for setting the default VRM model
	router.HandleFunc("/api/write/model/set", func(w http.ResponseWriter, r *http.Request) {

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

	// Route for saving the internal state of the scene config
	router.HandleFunc("/api/write/config/scene/set", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to save current scene")

		// Access control
		allowHTTPAllPerms(&w)

		if err := saveScene(sceneCfg, appCfg.SceneFilePath); err != nil {
			log.Println(err)
			return
		}

	}).Methods("PUT", "OPTIONS")

	// Route for retrieving the initial config for the server
	router.HandleFunc("/api/read/config/app/get", func(w http.ResponseWriter, r *http.Request) {

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

	// Route for returning the names of all available MotionReceiver sources
	router.HandleFunc("/api/read/receiver/get", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to list all available motion receivers")
		var receiverNames []string

		for name := range receiverMap {
			receiverNames = append(receiverNames, name)
		}

		bytes, err := json.Marshal(receiverNames)
		if err != nil {
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)

	}).Methods("GET", "OPTIONS")

	// Route for changing the active MotionReceiver source used
	router.HandleFunc("/api/write/receiver/set", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to change the MotionReceiver source for model")

		// Read in the request body into bytes, cast to string
		reqBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			return
		}
		suggestedReceiver := string(reqBytes)

		// Error if the suggested receiver does not exist
		if receiverMap[suggestedReceiver] == nil {
			log.Printf("\"%s\" does not exist!", suggestedReceiver)
			return
		}

		// Switch the active receiver
		activeReceiver = receiverMap[suggestedReceiver]
		log.Printf("Successfully changed the active receiver to %s", suggestedReceiver)

	}).Methods("PUT", "OPTIONS")

	// All other requests are sent to the embedded web frontend
	frontendRoot, err := frontend.FS()
	if err != nil {
		log.Fatal(err)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(frontendRoot)))

	return router

}
