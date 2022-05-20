package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hypebeast/go-osc/osc"
)

var (
	liveVRM    = vrmType{}    // VRM transformation data, updated from sources
	liveCamera = cameraType{} // Camera transformation data, updated from all client
	wsPool     = make(map[string]*websocket.Conn)

	modelUpdateFrequency  = 60 // Times per second VRM model data is sent to a client
	cameraUpdateFrequency = 1  // Times per second camera transformation data is sent to all clients
)

type cameraType struct {
	Position objPosition `json:"position,omitempty"`
	Target   objPosition `json:"target,omitempty"`
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
	EyeBlinkLeft        float32 `json:"eye_blink_left,omitempty"`
	EyeLookDownLeft     float32 `json:"eye_look_down_left,omitempty"`
	EyeLookInLeft       float32 `json:"eye_look_in_left,omitempty"`
	EyeLookOutLeft      float32 `json:"eye_look_out_left,omitempty"`
	EyeLookUpLeft       float32 `json:"eye_look_up_left,omitempty"`
	EyeSquintLeft       float32 `json:"eye_squint_left,omitempty"`
	EyeWideLeft         float32 `json:"eye_wide_left,omitempty"`
	EyeBlinkRight       float32 `json:"eye_blink_right,omitempty"`
	EyeLookDownRight    float32 `json:"eye_look_down_right,omitempty"`
	EyeLookInRight      float32 `json:"eye_look_in_right,omitempty"`
	EyeLookOutRight     float32 `json:"eye_look_out_right,omitempty"`
	EyeLookUpRight      float32 `json:"eye_look_up_right,omitempty"`
	EyeSquintRight      float32 `json:"eye_squint_right,omitempty"`
	EyeWideRight        float32 `json:"eye_wide_right,omitempty"`
	JawForward          float32 `json:"jaw_forward,omitempty"`
	JawLeft             float32 `json:"jaw_left,omitempty"`
	JawRight            float32 `json:"jaw_right,omitempty"`
	JawOpen             float32 `json:"jaw_open,omitempty"`
	MouthClose          float32 `json:"mouth_close,omitempty"`
	MouthFunnel         float32 `json:"mouth_funnel,omitempty"`
	MouthPucker         float32 `json:"mouth_pucker,omitempty"`
	MouthLeft           float32 `json:"mouth_left,omitempty"`
	MouthRight          float32 `json:"mouth_right,omitempty"`
	MouthSmileLeft      float32 `json:"mouth_smile_left,omitempty"`
	MouthSmileRight     float32 `json:"mouth_smile_right,omitempty"`
	MouthFrownLeft      float32 `json:"mouth_frown_left,omitempty"`
	MouthFrownRight     float32 `json:"mouth_frown_right,omitempty"`
	MouthDimpleLeft     float32 `json:"mouth_dimple_left,omitempty"`
	MouthDimpleRight    float32 `json:"mouth_dimple_right,omitempty"`
	MouthStretchLeft    float32 `json:"mouth_stretch_left,omitempty"`
	MouthStretchRight   float32 `json:"mouth_stretch_right,omitempty"`
	MouthRollLower      float32 `json:"mouth_roll_lower,omitempty"`
	MouthRollUpper      float32 `json:"mouth_roll_upper,omitempty"`
	MouthShrugLower     float32 `json:"mouth_shrug_lower,omitempty"`
	MouthShrugUpper     float32 `json:"mouth_shrug_upper,omitempty"`
	MouthPressLeft      float32 `json:"mouth_press_left,omitempty"`
	MouthPressRight     float32 `json:"mouth_press_right,omitempty"`
	MouthLowerDownLeft  float32 `json:"mouth_lower_down_left,omitempty"`
	MouthLowerDownRight float32 `json:"mouth_lower_down_right,omitempty"`
	MouthUpperUpLeft    float32 `json:"mouth_upper_up_left,omitempty"`
	MouthUpperUpRight   float32 `json:"mouth_upper_up_right,omitempty"`
	BrowDownLeft        float32 `json:"brow_down_left,omitempty"`
	BrowDownRight       float32 `json:"brow_down_right,omitempty"`
	BrowInnerUp         float32 `json:"brow_inner_up,omitempty"`
	BrowOuterUpLeft     float32 `json:"brow_outer_up_left,omitempty"`
	BrowOuterUpRight    float32 `json:"brow_outer_up_right,omitempty"`
	CheekPuff           float32 `json:"cheek_puff,omitempty"`
	CheekSquintLeft     float32 `json:"cheek_squint_left,omitempty"`
	CheekSquintRight    float32 `json:"cheek_squint_right,omitempty"`
	NoseSneerLeft       float32 `json:"nose_sneer_left,omitempty"`
	NoseSneerRight      float32 `json:"nose_sneer_right,omitempty"`
	TongueOut           float32 `json:"tongue_out,omitempty"`
}

// Object positioning properties for any given object
type objPosition struct {
	X float32 `json:"x,omitempty"`
	Y float32 `json:"y,omitempty"`
	Z float32 `json:"z,omitempty"`
}

// Quaternion rotation properties for any given object
type objQuaternionRotation struct {
	X float32 `json:"x,omitempty"`
	Y float32 `json:"y,omitempty"`
	Z float32 `json:"z,omitempty"`
	W float32 `json:"w,omitempty"`
}

type objSphericalRotation struct {
	AzimuthAngle float32 `json:"azimuth,omitempty"`
	PolarAngle   float32 `json:"polar,omitempty"`
}

// TODO: add Euler rotation alternative to Quaternion rotations. Math might be involved...
type boneRotation struct {
	Quaternion objQuaternionRotation `json:"quaternion,omitempty"`
	Spherical  objSphericalRotation  `json:"spherical,omitempty"`
	//Euler eulerRotation `json:"euler,omitempty"`
}

// Properties of a single VRM vrmBone
type vrmBone struct {
	Position objPosition  `json:"position,omitempty"`
	Rotation boneRotation `json:"rotation,omitempty"`
}

// All bones used in a VRM model, based off of Unity's HumanBodyBones
type vrmBones struct {
	TongueOut               float32 `json:"tongue_out"`
	Hips                    vrmBone `json:"hips"`
	LeftUpperLeg            vrmBone `json:"left_upper_leg"`
	RightUpperLeg           vrmBone `json:"right_upper_leg"`
	LeftLowerLeg            vrmBone `json:"left_lower_leg"`
	RightLowerLeg           vrmBone `json:"right_lower_leg"`
	LeftFoot                vrmBone `json:"left_foot"`
	RightFoot               vrmBone `json:"right_foot"`
	Spine                   vrmBone `json:"spine"`
	Chest                   vrmBone `json:"chest"`
	UpperChest              vrmBone `json:"upper_chest"`
	Neck                    vrmBone `json:"neck"`
	Head                    vrmBone `json:"head"`
	LeftShoulder            vrmBone `json:"left_shoulder"`
	RightShoulder           vrmBone `json:"right_shoulder"`
	LeftUpperArm            vrmBone `json:"left_upper_arm"`
	RightUpperArm           vrmBone `json:"right_upper_arm"`
	LeftLowerArm            vrmBone `json:"left_lower_arm"`
	RightLowerArm           vrmBone `json:"right_lower_arm"`
	LeftHand                vrmBone `json:"left_hand"`
	RightHand               vrmBone `json:"right_hand"`
	LeftToes                vrmBone `json:"left_toes"`
	RightToes               vrmBone `json:"right_toes"`
	LeftEye                 vrmBone `json:"left_eye"`
	RightEye                vrmBone `json:"right_eye"`
	Jaw                     vrmBone `json:"jaw"`
	LeftThumbProximal       vrmBone `json:"left_thumb_proximal"`
	LeftThumbIntermediate   vrmBone `json:"left_thumb_intermediate"`
	LeftThumbDistal         vrmBone `json:"left_thumb_distal"`
	LeftIndexProximal       vrmBone `json:"left_index_proximal"`
	LeftIndexIntermediate   vrmBone `json:"left_index_intermediate"`
	LeftIndexDistal         vrmBone `json:"left_index_distal"`
	LeftMiddleProximal      vrmBone `json:"left_middle_proximal"`
	LeftMiddleIntermediate  vrmBone `json:"left_middle_intermediate"`
	LeftMiddleDistal        vrmBone `json:"left_middle_distal"`
	LeftRingProximal        vrmBone `json:"left_ring_proximal"`
	LeftRingIntermediate    vrmBone `json:"left_ring_intermediate"`
	LeftRingDistal          vrmBone `json:"left_ring_distal"`
	LeftLittleProximal      vrmBone `json:"left_little_proximal"`
	LeftLittleIntermediate  vrmBone `json:"left_little_intermediate"`
	LeftLittleDistal        vrmBone `json:"left_little_distal"`
	RightThumbProximal      vrmBone `json:"right_thumb_proximal"`
	RightThumbIntermediate  vrmBone `json:"right_thumb_intermediate"`
	RightThumbDistal        vrmBone `json:"right_thumb_distal"`
	RightIndexProximal      vrmBone `json:"right_index_proximal"`
	RightIndexIntermediate  vrmBone `json:"right_index_intermediate"`
	RightIndexDistal        vrmBone `json:"right_index_distal"`
	RightMiddleProximal     vrmBone `json:"right_middle_proximal"`
	RightMiddleIntermediate vrmBone `json:"right_middle_intermediate"`
	RightMiddleDistal       vrmBone `json:"right_middle_distal"`
	RightRingProximal       vrmBone `json:"right_ring_proximal"`
	RightRingIntermediate   vrmBone `json:"right_ring_intermediate"`
	RightRingDistal         vrmBone `json:"right_ring_distal"`
	RightLittleProximal     vrmBone `json:"right_little_proximal"`
	RightLittleIntermediate vrmBone `json:"right_little_intermediate"`
	RightLittleDistal       vrmBone `json:"right_little_distal"`
	LastBone                vrmBone `json:"last_bone"`
}

// Helper function to convert CamelCase string to snake_case
func camelToSnake(str string) (string, error) {
	matchFirstCap, err := regexp.Compile("(.)([A-Z][a-z]+)")
	if err != nil {
		return "", err
	}

	matchAllCap, err := regexp.Compile("([a-z0-9])([A-Z])")
	if err != nil {
		return "", err
	}

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake), nil

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

// Get first index value of an OSC message, which is the key in the VMC protocol specification
// Note that, specifically in the VMC protcol, all key names are in CamelCase
// This is not ideal for javascript naming conventions...or maybe I don't know what I'm doing
// and am just adding too much excess code...
func parseKey(msg *osc.Message) (string, error) {

	rawKey, ok := msg.Arguments[0].(string)
	if !ok {
		return "", fmt.Errorf("Unable to type assert OSC message string key: %s", msg)
	}

	key, err := camelToSnake(rawKey)
	if err != nil {
		return "", err
	}

	return key, nil

}

// Listen for face and bone data through OSC from a device in the VMC protocol format
func listenVMC(address string, port int) {

	d := osc.NewStandardDispatcher()

	// Now to add whatever routes are needed, according to the VMC spec

	d.AddMsgHandler("/VMC/Ext/Blend/Val", func(msg *osc.Message) {

	})

	// Bone position and rotation request handler
	d.AddMsgHandler("/VMC/Ext/Bone/Pos", func(msg *osc.Message) {

		key, err := parseKey(msg)
		if err != nil {
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

// Update the camera position for all clients
func updateClientCameras(wsConns map[string]*websocket.Conn) {
	log.Println(wsConns)

	for k, v := range wsConns {

		log.Printf("Attempting to update camera for client %s...", v.RemoteAddr())
		// If unable to write JSON to one of the WebSockets, close it and forget reference of it in our map
		if err := v.WriteJSON(liveCamera); err != nil {
			log.Println()
			log.Printf("Error writing JSON to camera client, closing %s\n", v.RemoteAddr())
			v.Close()
			delete(wsConns, k)
			continue

		}

		log.Println("Success!")

	}

}

// Entrypoint
func main() {

	// Background listen and serve for face and bone data
	go listenVMC("0.0.0.0", 39540)

	router := mux.NewRouter()

	// Route for relaying the internal state of the camera to all clients
	router.HandleFunc("/api/camera", func(w http.ResponseWriter, r *http.Request) {

		log.Printf("Received camera WebSocket request from %s\n", r.RemoteAddr)

		ws, err := wsUpgrade(w, r)
		if err != nil {
			log.Println(err)
		}

		// Store reference to this WebSocket in our pool of clients, so all are updated at once
		wsID := strconv.Itoa(len(wsPool) + 1)
		wsPool[wsID] = ws

		// For every time the WebSocket is open, decode JSON camera positioning from request, update liveCamera with it
		for {

			if err := ws.ReadJSON(&liveCamera); err != nil {
				log.Printf("Error reading JSON from camera client, closing %s\n", ws.RemoteAddr())
				ws.Close()
				return

			}

			log.Printf("Received update request from client %s\n", ws.RemoteAddr())
			updateClientCameras(wsPool)

		}

	})

	router.HandleFunc("/api/model", func(w http.ResponseWriter, r *http.Request) {

		log.Printf("Received model API request from %s\n", r.RemoteAddr)

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
			time.Sleep(time.Duration(1e9 / modelUpdateFrequency))

		}

	})

	// Blocking listen and serve for WebSockets and API server
	http.ListenAndServe("127.0.0.1:3579", router)

}
