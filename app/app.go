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
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hypebeast/go-osc/osc"
	"github.com/thatpix3l/fntwo/cfg"
	"github.com/thatpix3l/fntwo/frontend"
	"github.com/thatpix3l/fntwo/obj"
)

var (
	liveVRM  = obj.VRM{}   // VRM transformation data, updated from sources
	initCfg  *cfg.Initial  // Initial config for settings of the app
	sceneCfg = cfg.Scene{} // Scene config for various live data
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
		log.Print("Received message from broadcasting channel")
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

// Assuming everything after the first index is bone data, type assert it as a slice of float32
// The positioning of the data is special, where the index is as follows:
// index 0, 1, 2: bone position X, Y, Z
// index 3, 4, 5, 6: bone quaternion rotation X, Y, Z, W
func parseBone(msg *osc.Message) ([]float32, error) {
	var boneData []float32
	for _, v := range msg.Arguments[1:] {
		coord, ok := v.(float32)
		if !ok {
			return nil, fmt.Errorf("Unable to type assert OSC message as []float32 bone coords: %s", msg)
		}

		boneData = append(boneData, coord)

	}

	return boneData, nil

}

// Listen for face and bone data through OSC from a device in the VMC protocol format
func listenVMC(address string, port int) {

	log.Printf("Listening for VMC model transformation data on %s:%d", address, port)

	d := osc.NewStandardDispatcher()

	// Now to add whatever routes are needed, according to the VMC spec

	// BlendShapes handler
	d.AddMsgHandler("/VMC/Ext/Blend/Val", func(msg *osc.Message) {

		// Get key name
		key, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}

		// Get value float32
		blendValue, ok := msg.Arguments[1].(float32)
		if !ok {
			return
		}

		// Set max and min for blendValue to betweem 0 and 1
		if blendValue > 1 {
			blendValue = 1
		}

		if blendValue < 0 {
			blendValue = 0
		}

		// Create map structure, containing a single key with a single value
		newMap := make(map[string]float32)
		newMap[key] = blendValue

		mapBytes, err := json.Marshal(newMap)
		if err != nil {
			return
		}

		if err := json.Unmarshal(mapBytes, &liveVRM.BlendShapes.Face); err != nil {
			return
		}

	})

	// Bone position and rotation request handler
	d.AddMsgHandler("/VMC/Ext/Bone/Pos", func(msg *osc.Message) {

		// Bone name
		key, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}

		// Bone position and rotation
		value, err := parseBone(msg)
		if err != nil {
			return
		}

		// Map with bone name string keys and obj.Bone values
		newBoneMap := make(map[string]obj.Bone)

		// New bone structure
		newBone := obj.Bone{
			Position: obj.Position{
				X: value[0],
				Y: value[1],
				Z: value[2],
			},
			Rotation: obj.QuaternionRotation{
				X: value[3],
				Y: value[4],
				Z: value[5],
				W: value[6],
			},
		}

		// The the bone with the name from the OSC message will have this new bone data
		newBoneMap[key] = newBone

		// Marshal our map representation into bytes
		newBoneBytes, err := json.Marshal(newBoneMap)
		if err != nil {
			log.Println(err)
			return

		}

		// Finally, unmarshal the JSON representation of our bones into the bones section of our VRM
		if err := json.Unmarshal(newBoneBytes, &liveVRM.Bones); err != nil {
			log.Println(err)
			return
		}

	})

	// OSC server configuration
	addr := address + ":" + strconv.Itoa(port)
	server := &osc.Server{
		Addr:       addr,
		Dispatcher: d,
	}

	// Blocking listen and serve
	server.ListenAndServe()

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

// Entrypoint
func Start(initialConfig *cfg.Initial) {

	// Store pointer of generated config file to use throughout this program
	initCfg = initialConfig

	// Load scene config from disk, if it even exists
	if err := loadScene(initialConfig.SceneCfgFile); err != nil {
		log.Println(err)
	}

	// Background listen and serve for face and bone data
	go listenVMC(initCfg.VmcListenIP, initCfg.VmcListenPort)

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
			time.Sleep(time.Duration(1e9 / initCfg.ModelUpdateFrequency))

		}

	})

	// Route for getting the default VRM model
	router.HandleFunc("/api/model", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to retrieve default VRM file")

		// Set model name and CORS policy
		w.Header().Set("Content-Disposition", "attachment; filename=default.vrm")
		allowHTTPAllPerms(&w)

		// Serve default VRM file
		http.ServeFile(w, r, initialConfig.VRMFile)

	}).Methods("GET")

	// Route for setting the default VRM model
	router.HandleFunc("/api/model", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to set default VRM file")

		allowHTTPAllPerms(&w)

		// Destination VRM file on system
		dest, err := os.Create(initCfg.VRMFile)
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
	router.HandleFunc("/api/scene", func(w http.ResponseWriter, r *http.Request) {

		log.Print("Received request to save current scene")

		// Access control
		allowHTTPAllPerms(&w)

		// Convert the scene config in memory into bytes
		sceneCfgBytes, err := json.MarshalIndent(sceneCfg, "", " ")
		if err != nil {
			log.Print(err)
			return
		}

		// Store config bytes into file
		if err := os.WriteFile(initialConfig.SceneCfgFile, sceneCfgBytes, 0644); err != nil {
			log.Print(err)
			return
		}

	}).Methods("PUT")

	// HTTP route for retrieving the initial config for the server
	router.HandleFunc("/api/initialConfig", func(w http.ResponseWriter, r *http.Request) {

		log.Println("Received request to retrieve initial config")

		// Access control
		allowHTTPAllPerms(&w)

		// Marshal initial config into bytes
		initCfgBytes, err := json.Marshal(initCfg)
		if err != nil {
			log.Println(err)
			return
		}

		// Reply back to request with byte-format of initial config
		w.Header().Set("Content-Type", "application/json")
		w.Write(initCfgBytes)

	}).Methods("GET")

	// All other requests are sent to the embedded web frontend
	frontendRoot, err := frontend.FS()
	if err != nil {
		log.Fatal(err)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(frontendRoot)))

	// Blocking listen and serve for WebSockets and API server
	log.Printf("Serving frontend and listening for clients/API queries on %s", initCfg.GetWebServerAddress())
	http.ListenAndServe(initCfg.GetWebServerAddress(), router)

}
