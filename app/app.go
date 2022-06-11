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
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/frontend"
	"github.com/thatpix3l/fntwo/obj"
	"github.com/thatpix3l/fntwo/receivers/facemotion3d"
	"github.com/thatpix3l/fntwo/receivers/virtualmotioncapture"
)

var (
	liveVRM  = obj.VRM{}      // VRM transformation data, updated from sources
	appCfg   *config.App      // Initial config for settings of the app
	sceneCfg = config.Scene{} // Scene config for various live data
)

type websocketPool struct {
	Clients          map[string]websocketClient // Map of websocketClient
	BroadcastChannel chan obj.Camera            // Each websocketClient has access to the exact same channel for any camera data
}

type websocketClient struct {
	ID      string          // A generally unique, runtime ID for each websocketClient
	Channel chan obj.Camera // Channel for listening to any new camera data and relaying to client, which is updated by other clients
}

// Create and return a new websocketPool
func newPool() websocketPool {
	p := websocketPool{
		Clients:          make(map[string]websocketClient),
		BroadcastChannel: make(chan obj.Camera),
	}
	return p
}

// Relay the given camera data to the broadcasting channel, which will eventually propagate to all other clients
func (p websocketPool) send(msg obj.Camera) {

	// Store copy of the camera data
	sceneCfg.Camera = obj.Camera{
		GazeTowards: obj.Position{
			X: msg.GazeTowards.X,
			Y: msg.GazeTowards.Y,
			Z: msg.GazeTowards.Z,
		},
		GazeFrom: obj.Position{
			X: msg.GazeFrom.X,
			Y: msg.GazeFrom.Y,
			Z: msg.GazeFrom.Z,
		},
	}

	// Relay message to broadcast channel
	p.BroadcastChannel <- msg
}

// Get the current count of all clients
func (p websocketPool) count() int {
	return len(p.Clients)
}

// Log the number of connected clients
func (p websocketPool) logCount() {
	log.Printf("Number of connected clients: %d", p.count())
}

// Add a new websocketClient with the the given ID
func (p websocketPool) add(id string) {

	log.Printf("Adding WebSocket camera client with ID %s", id)

	p.Clients[id] = websocketClient{
		ID:      id,
		Channel: make(chan obj.Camera),
	}

}

// Remove a websocketClient with the given ID
func (p websocketPool) remove(id string) {

	log.Printf("Removing WebSocket client with ID %s", id)

	close(p.Clients[id].Channel)
	delete(p.Clients, id)

}

// Start listening for new camera data from the broadcast channel
func (p websocketPool) start() {

	log.Printf("Listening for messages from broadcasting channel")

	for {
		msg := <-p.BroadcastChannel
		for _, client := range p.Clients {
			log.Printf("Updating client %s", client.ID)
			client.Channel <- msg
		}
	}

}

// Listening for camera data from frontend and relay to broadcasting channel.
// Simultaneously, relay from backend to frontend if new camera data has been updated from other clients
func (p websocketPool) listen(id string, ws *websocket.Conn) {

	// On first time connect, send to client the scene config
	ws.WriteJSON(sceneCfg.Camera)

	// Background listen for broadcast messages from channel with this ID
	go func() {
		for {
			data, ok := <-p.Clients[id].Channel
			if !ok {
				return
			}
			ws.WriteJSON(data)
		}
	}()

	// For each time a valid JSON request is received, decode it and send it down the message channel
	for {

		var camera obj.Camera
		if err := ws.ReadJSON(&camera); err != nil {
			p.remove(id)
			ws.Close()
			return

		}

		log.Printf("Client %s from %s sent new camera data", id, ws.RemoteAddr())
		p.send(camera)

	}

}

// Helper function to generate a random string
func randomString(n int) string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

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

// Helper func to load a scene config file
func loadScene(sceneCfgPath string) error {

	// Read in a scene JSON config file
	content, err := os.ReadFile(sceneCfgPath)
	if err != nil {
		return err
	}

	// Unmarshal into memory's scene config
	if err := json.Unmarshal(content, &sceneCfg); err != nil {
		return err
	}

	return nil

}

// Helper func to allow all origin, headers, and methods for HTTP requests.
func allowHTTPAllPerms(wPtr *http.ResponseWriter) {

	// Set CORS policy
	(*wPtr).Header().Set("Access-Control-Allow-Origin", "*")
	(*wPtr).Header().Set("Access-Control-Allow-Methods", "*")
	(*wPtr).Header().Set("Access-Control-Allow-Headers", "*")

}

func saveScene(scene config.Scene) error {

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
	defautlSceneBytes, err := json.MarshalIndent(defaultScene, "", "    ")
	if err != nil {
		return err
	}

	// Write to file the scene data
	if _, err := sceneFile.Write(defautlSceneBytes); err != nil {
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
		log.Print(err)
	}

	// Load scene config from disk
	if err := loadScene(appCfg.SceneFilePath); err != nil {
		log.Println(err)
	}

	// Background listen and serve for face and bone tracking
	vmcServer := virtualmotioncapture.New(&liveVRM, appCfg)
	fm3dServer := facemotion3d.New(&liveVRM, appCfg)
	go vmcServer.Start()
	go fm3dServer.Start()

	// Create new WebSocket pool, listen in background for messages
	wsPool := newPool()
	go wsPool.start()

	router := mux.NewRouter()

	// Route for relaying the internal state of the camera to all clients
	router.HandleFunc("/client/camera", func(w http.ResponseWriter, r *http.Request) {

		// Upgrade GET request to WebSocket
		ws, err := wsUpgrade(w, r)
		if err != nil {
			log.Println(err)
		}

		// Unique identifier for this WebSocket session
		wsID := randomString(6)

		// Create new client with this WebSocket connection and indentifier
		wsPool.add(wsID)

		// Log count of clients before listening
		wsPool.logCount()

		// Blocking listen
		wsPool.listen(wsID, ws)

		// Log count of clients after done listening
		wsPool.logCount()

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

			if err := ws.WriteJSON(liveVRM); err != nil {
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

		log.Print("Received request to save current scene")

		// Access control
		allowHTTPAllPerms(&w)

		if err := saveScene(sceneCfg); err != nil {
			log.Print(err)
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

	// Blocking listen and serve for WebSockets and API server
	log.Printf("Serving frontend and listening for clients/API queries on %s", appCfg.GetWebServerAddress())
	http.ListenAndServe(appCfg.GetWebServerAddress(), router)

}
