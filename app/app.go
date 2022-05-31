package app

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hypebeast/go-osc/osc"
	"github.com/thatpix3l/fntwo/cfg"
)

var (
	liveVRM = vrmType{} // VRM transformation data, updated from sources
	config  *cfg.Keys   // Final config file from command-line usage
)

type websocketPool struct {
	Clients          map[string]websocketClient
	BroadcastChannel chan cameraType
}

type websocketClient struct {
	ID      string
	Channel chan cameraType
}

func newPool() websocketPool {
	p := websocketPool{
		Clients:          make(map[string]websocketClient),
		BroadcastChannel: make(chan cameraType),
	}
	return p
}

func (p websocketPool) send(msg cameraType) {
	p.BroadcastChannel <- msg
}

func (p websocketPool) count() int {
	return len(p.Clients)
}

func (p websocketPool) add(id string) {
	log.Printf("Adding WebSocket client with ID %s", id)
	newClient := websocketClient{
		ID:      id,
		Channel: make(chan cameraType),
	}

	p.Clients[id] = newClient

}

func (p websocketPool) remove(id string) {
	log.Printf("Removing WebSocket client with ID %s", id)
	close(p.Clients[id].Channel)
	delete(p.Clients, id)
}

func (p websocketPool) start() {
	log.Printf("Listening for messages from broadcasting channel")
	for {
		msg := <-p.BroadcastChannel
		log.Printf("Received message from broadcasting channel")
		for _, client := range p.Clients {
			log.Printf("Updating client %s", client.ID)
			client.Channel <- msg
		}
	}
}

func (p websocketPool) listen(id string, ws *websocket.Conn) {

	// Add new client with ID
	p.add(id)
	log.Printf("Pool count of clients: %d", p.count())

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

		var camera cameraType
		if err := ws.ReadJSON(&camera); err != nil {
			log.Printf("Pool count of clients: %d", p.count())
			p.remove(id)
			ws.Close()
			return

		}

		log.Printf("Client %s from %s sent new camera data", id, ws.RemoteAddr())
		p.send(camera)

	}

}

type cameraType struct {
	Position objPosition `json:"position"`
	Target   objPosition `json:"target"`
}

// Entire data related to transformations for a VRM model
type vrmType struct {
	Bones       vrmBones       `json:"bones,omitempty"`        // Updated bone data
	BlendShapes vrmBlendShapes `json:"blend_shapes,omitempty"` // Updated blend shape data
}

// All available VRM blend shapes
type vrmBlendShapes struct {
	FaceBlendShapes vrmFaceBlendShapes `json:"face,omitempty"`
}

// The available face blend shapes to modify, based off of Apple's 52 BlendShape AR-kit spec
type vrmFaceBlendShapes struct {
	EyeBlinkLeft        float32 `json:"EyeBlinkLeft"`
	EyeLookDownLeft     float32 `json:"EyeLookDownLeft"`
	EyeLookInLeft       float32 `json:"EyeLookInLeft"`
	EyeLookOutLeft      float32 `json:"EyeLookOutLeft"`
	EyeLookUpLeft       float32 `json:"EyeLookUpLeft"`
	EyeSquintLeft       float32 `json:"EyeSquintLeft"`
	EyeWideLeft         float32 `json:"EyeWideLeft"`
	EyeBlinkRight       float32 `json:"EyeBlinkRight"`
	EyeLookDownRight    float32 `json:"EyeLookDownRight"`
	EyeLookInRight      float32 `json:"EyeLookInRight"`
	EyeLookOutRight     float32 `json:"EyeLookOutRight"`
	EyeLookUpRight      float32 `json:"EyeLookUpRight"`
	EyeSquintRight      float32 `json:"EyeSquintRight"`
	EyeWideRight        float32 `json:"EyeWideRight"`
	JawForward          float32 `json:"JawForward"`
	JawLeft             float32 `json:"JawLeft"`
	JawRight            float32 `json:"JawRight"`
	JawOpen             float32 `json:"JawOpen"`
	MouthClose          float32 `json:"MouthClose"`
	MouthFunnel         float32 `json:"MouthFunnel"`
	MouthPucker         float32 `json:"MouthPucker"`
	MouthLeft           float32 `json:"MouthLeft"`
	MouthRight          float32 `json:"MouthRight"`
	MouthSmileLeft      float32 `json:"MouthSmileLeft"`
	MouthSmileRight     float32 `json:"MouthSmileRight"`
	MouthFrownLeft      float32 `json:"MouthFrownLeft"`
	MouthFrownRight     float32 `json:"MouthFrownRight"`
	MouthDimpleLeft     float32 `json:"MouthDimpleLeft"`
	MouthDimpleRight    float32 `json:"MouthDimpleRight"`
	MouthStretchLeft    float32 `json:"MouthStretchLeft"`
	MouthStretchRight   float32 `json:"MouthStretchRight"`
	MouthRollLower      float32 `json:"MouthRollLower"`
	MouthRollUpper      float32 `json:"MouthRollUpper"`
	MouthShrugLower     float32 `json:"MouthShrugLower"`
	MouthShrugUpper     float32 `json:"MouthShrugUpper"`
	MouthPressLeft      float32 `json:"MouthPressLeft"`
	MouthPressRight     float32 `json:"MouthPressRight"`
	MouthLowerDownLeft  float32 `json:"MouthLowerDownLeft"`
	MouthLowerDownRight float32 `json:"MouthLowerDownRight"`
	MouthUpperUpLeft    float32 `json:"MouthUpperUpLeft"`
	MouthUpperUpRight   float32 `json:"MouthUpperUpRight"`
	BrowDownLeft        float32 `json:"BrowDownLeft"`
	BrowDownRight       float32 `json:"BrowDownRight"`
	BrowInnerUp         float32 `json:"BrowInnerUp"`
	BrowOuterUpLeft     float32 `json:"BrowOuterUpLeft"`
	BrowOuterUpRight    float32 `json:"BrowOuterUpRight"`
	CheekPuff           float32 `json:"CheekPuff"`
	CheekSquintLeft     float32 `json:"CheekSquintLeft"`
	CheekSquintRight    float32 `json:"CheekSquintRight"`
	NoseSneerLeft       float32 `json:"NoseSneerLeft"`
	NoseSneerRight      float32 `json:"NoseSneerRight"`
	TongueOut           float32 `json:"TongueOut"`
}

// Object positioning properties for any given object
type objPosition struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// Quaternion rotation properties for any given object
type objQuaternionRotation struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
	W float32 `json:"w"`
}

type objSphericalRotation struct {
	AzimuthAngle float32 `json:"azimuth"`
	PolarAngle   float32 `json:"polar"`
}

// TODO: add Euler rotation alternative to Quaternion rotations. Math might be involved...
type boneRotation struct {
	Quaternion objQuaternionRotation `json:"quaternion"`
	Spherical  objSphericalRotation  `json:"spherical,omitempty"`
	//Euler eulerRotation `json:"euler,omitempty"`
}

// Properties of a single VRM vrmBone
type vrmBone struct {
	Position objPosition  `json:"position"`
	Rotation boneRotation `json:"rotation"`
}

// All bones used in a VRM model, based off of Unity's HumanBodyBones
type vrmBones struct {
	Hips                    vrmBone `json:"Hips"`
	LeftUpperLeg            vrmBone `json:"LeftUpperLeg"`
	RightUpperLeg           vrmBone `json:"RightUpperLeg"`
	LeftLowerLeg            vrmBone `json:"LeftLowerLeg"`
	RightLowerLeg           vrmBone `json:"RightLowerLeg"`
	LeftFoot                vrmBone `json:"LeftFoot"`
	RightFoot               vrmBone `json:"RightFoot"`
	Spine                   vrmBone `json:"Spine"`
	Chest                   vrmBone `json:"Chest"`
	UpperChest              vrmBone `json:"UpperChest"`
	Neck                    vrmBone `json:"Neck"`
	Head                    vrmBone `json:"Head"`
	LeftShoulder            vrmBone `json:"LeftShoulder"`
	RightShoulder           vrmBone `json:"RightShoulder"`
	LeftUpperArm            vrmBone `json:"LeftUpperArm"`
	RightUpperArm           vrmBone `json:"RightUpperArm"`
	LeftLowerArm            vrmBone `json:"LeftLowerArm"`
	RightLowerArm           vrmBone `json:"RightLowerArm"`
	LeftHand                vrmBone `json:"LeftHand"`
	RightHand               vrmBone `json:"RightHand"`
	LeftToes                vrmBone `json:"LeftToes"`
	RightToes               vrmBone `json:"RightToes"`
	LeftEye                 vrmBone `json:"LeftEye"`
	RightEye                vrmBone `json:"RightEye"`
	Jaw                     vrmBone `json:"Jaw"`
	LeftThumbProximal       vrmBone `json:"LeftThumbProximal"`
	LeftThumbIntermediate   vrmBone `json:"LeftThumbIntermediate"`
	LeftThumbDistal         vrmBone `json:"LeftThumbDistal"`
	LeftIndexProximal       vrmBone `json:"LeftIndexProximal"`
	LeftIndexIntermediate   vrmBone `json:"LeftIndexIntermediate"`
	LeftIndexDistal         vrmBone `json:"LeftIndexDistal"`
	LeftMiddleProximal      vrmBone `json:"LeftMiddleProximal"`
	LeftMiddleIntermediate  vrmBone `json:"LeftMiddleIntermediate"`
	LeftMiddleDistal        vrmBone `json:"LeftMiddleDistal"`
	LeftRingProximal        vrmBone `json:"LeftRingProximal"`
	LeftRingIntermediate    vrmBone `json:"LeftRingIntermediate"`
	LeftRingDistal          vrmBone `json:"LeftRingDistal"`
	LeftLittleProximal      vrmBone `json:"LeftLittleProximal"`
	LeftLittleIntermediate  vrmBone `json:"LeftLittleIntermediate"`
	LeftLittleDistal        vrmBone `json:"LeftLittleDistal"`
	RightThumbProximal      vrmBone `json:"RightThumbProximal"`
	RightThumbIntermediate  vrmBone `json:"RightThumbIntermediate"`
	RightThumbDistal        vrmBone `json:"RightThumbDistal"`
	RightIndexProximal      vrmBone `json:"RightIndexProximal"`
	RightIndexIntermediate  vrmBone `json:"RightIndexIntermediate"`
	RightIndexDistal        vrmBone `json:"RightIndexDistal"`
	RightMiddleProximal     vrmBone `json:"RightMiddleProximal"`
	RightMiddleIntermediate vrmBone `json:"RightMiddleIntermediate"`
	RightMiddleDistal       vrmBone `json:"RightMiddleDistal"`
	RightRingProximal       vrmBone `json:"RightRingProximal"`
	RightRingIntermediate   vrmBone `json:"RightRingIntermediate"`
	RightRingDistal         vrmBone `json:"RightRingDistal"`
	RightLittleProximal     vrmBone `json:"RightLittleProximal"`
	RightLittleIntermediate vrmBone `json:"RightLittleIntermediate"`
	RightLittleDistal       vrmBone `json:"RightLittleDistal"`
	LastBone                vrmBone `json:"LastBone"`
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

	log.Printf("Listening for VMC model data on %s:%d", address, port)

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

		if err := json.Unmarshal(mapBytes, &liveVRM.BlendShapes.FaceBlendShapes); err != nil {
			return
		}

	})

	// Bone position and rotation request handler
	d.AddMsgHandler("/VMC/Ext/Bone/Pos", func(msg *osc.Message) {

		key, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}

		value, err := parseBone(msg)
		if err != nil {
			return
		}

		// Store bone data from OSC message into a map, containing one bone name with data
		// We're basically creating this structure:
		//
		// {
		//     "vrm_bone_name": {
		//         "position": {
		//             "x": bone_pos_x
		//             "y": bone_pos_y
		//             "z": bone_pos_z
		//         },
		//         "rotation": {
		//             "quaternion": {
		//                 "x": bone_rot_quat_x
		//                 "y": bone_rot_quat_y
		//                 "z": bone_rot_quat_z
		//                 "w": bone_rot_quat_w
		//             }
		//         }
		//     }
		// }

		newBones := make(map[string]vrmBone)

		newBone := vrmBone{
			Position: objPosition{
				X: value[0],
				Y: value[1],
				Z: value[2],
			},
			Rotation: boneRotation{
				Quaternion: objQuaternionRotation{
					X: value[3],
					Y: value[4],
					Z: value[5],
					W: value[6],
				},
			},
		}

		newBones[key] = newBone

		// Marshal our map representation of our bones data structure with one key changed, into bytes
		newBoneBytes, err := json.Marshal(newBones)
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

// Entrypoint
func Start(generatedConfig *cfg.Keys) {

	// Store pointer of generated config file to use throughout this program
	config = generatedConfig

	// Background listen and serve for face and bone data
	go listenVMC(config.VmcListenIP, config.VmcListenPort)

	// Create new WebSocket pool, listen in background for messages
	wsPool := newPool()
	go wsPool.start()

	router := mux.NewRouter()

	// Route for relaying the internal state of the camera to all clients
	router.HandleFunc("/api/camera", func(w http.ResponseWriter, r *http.Request) {

		// Upgrade GET request to WebSocket
		ws, err := wsUpgrade(w, r)
		if err != nil {
			log.Println(err)
		}

		// Unique identifier for this WebSocket session
		wsID := randomString(6)

		// Create new client with this WebSocket connection and indentifier
		wsPool.listen(wsID, ws)

	})

	// Live socket handler for updating VRM model data to all connections
	router.HandleFunc("/api/model", func(w http.ResponseWriter, r *http.Request) {

		log.Printf("Received model WebSocket request from %s", r.RemoteAddr)

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
			time.Sleep(time.Duration(1e9 / config.ModelUpdateFrequency))

		}

	})

	log.Printf("Listening for clients on %s", config.GetWebSocketAddress())

	// Blocking listen and serve for WebSockets and API server
	http.ListenAndServe(config.GetWebSocketAddress(), router)

}
