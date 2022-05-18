package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
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
	// Updated live from a source, as fast as possible
	liveVRM = vrmType{}
)

// Entire data related to transformations for a VRM model
type vrmType struct {
	Bones vrmBones `json:"bones,omit_empty"`
}

// Properties of a single VRM bone
type bone struct {
	PositionX float32 `json:"position_x,omit_empty"`
	PositionY float32 `json:"position_y,omit_empty"`
	PositionZ float32 `json:"position_z,omit_empty"`

	QuaternionX float32 `json:"quaternion_x,omit_empty"`
	QuaternionY float32 `json:"quaternion_y,omit_empty"`
	QuaternionZ float32 `json:"quaternion_z,omit_empty"`
	QuaternionW float32 `json:"quaternion_w,omit_empty"`
}

// All bones used in a VRM model, based off of Unity's HumanBodyBones
type vrmBones struct {
	TongueOut               float32 `json:"tongue_out,omit_empty"`
	Hips                    bone    `json:"hips,omit_empty"`
	LeftUpperLeg            bone    `json:"left_upper_leg,omit_empty"`
	RightUpperLeg           bone    `json:"right_upper_leg,omit_empty"`
	LeftLowerLeg            bone    `json:"left_lower_leg,omit_empty"`
	RightLowerLeg           bone    `json:"right_lower_leg,omit_empty"`
	LeftFoot                bone    `json:"left_foot,omit_empty"`
	RightFoot               bone    `json:"right_foot,omit_empty"`
	Spine                   bone    `json:"spine,omit_empty"`
	Chest                   bone    `json:"chest,omit_empty"`
	UpperChest              bone    `json:"upper_chest,omit_empty"`
	Neck                    bone    `json:"neck,omit_empty"`
	Head                    bone    `json:"head,omit_empty"`
	LeftShoulder            bone    `json:"left_shoulder,omit_empty"`
	RightShoulder           bone    `json:"right_shoulder,omit_empty"`
	LeftUpperArm            bone    `json:"left_upper_arm,omit_empty"`
	RightUpperArm           bone    `json:"right_upper_arm,omit_empty"`
	LeftLowerArm            bone    `json:"left_lower_arm,omit_empty"`
	RightLowerArm           bone    `json:"right_lower_arm,omit_empty"`
	LeftHand                bone    `json:"left_hand,omit_empty"`
	RightHand               bone    `json:"right_hand,omit_empty"`
	LeftToes                bone    `json:"left_toes,omit_empty"`
	RightToes               bone    `json:"right_toes,omit_empty"`
	LeftEye                 bone    `json:"left_eye,omit_empty"`
	RightEye                bone    `json:"right_eye,omit_empty"`
	Jaw                     bone    `json:"jaw,omit_empty"`
	LeftThumbProximal       bone    `json:"left_thumb_proximal,omit_empty"`
	LeftThumbIntermediate   bone    `json:"left_thumb_intermediate,omit_empty"`
	LeftThumbDistal         bone    `json:"left_thumb_distal,omit_empty"`
	LeftIndexProximal       bone    `json:"left_index_proximal,omit_empty"`
	LeftIndexIntermediate   bone    `json:"left_index_intermediate,omit_empty"`
	LeftIndexDistal         bone    `json:"left_index_distal,omit_empty"`
	LeftMiddleProximal      bone    `json:"left_middle_proximal,omit_empty"`
	LeftMiddleIntermediate  bone    `json:"left_middle_intermediate,omit_empty"`
	LeftMiddleDistal        bone    `json:"left_middle_distal,omit_empty"`
	LeftRingProximal        bone    `json:"left_ring_proximal,omit_empty"`
	LeftRingIntermediate    bone    `json:"left_ring_intermediate,omit_empty"`
	LeftRingDistal          bone    `json:"left_ring_distal,omit_empty"`
	LeftLittleProximal      bone    `json:"left_little_proximal,omit_empty"`
	LeftLittleIntermediate  bone    `json:"left_little_intermediate,omit_empty"`
	LeftLittleDistal        bone    `json:"left_little_distal,omit_empty"`
	RightThumbProximal      bone    `json:"right_thumb_proximal,omit_empty"`
	RightThumbIntermediate  bone    `json:"right_thumb_intermediate,omit_empty"`
	RightThumbDistal        bone    `json:"right_thumb_distal,omit_empty"`
	RightIndexProximal      bone    `json:"right_index_proximal,omit_empty"`
	RightIndexIntermediate  bone    `json:"right_index_intermediate,omit_empty"`
	RightIndexDistal        bone    `json:"right_index_distal,omit_empty"`
	RightMiddleProximal     bone    `json:"right_middle_proximal,omit_empty"`
	RightMiddleIntermediate bone    `json:"right_middle_intermediate,omit_empty"`
	RightMiddleDistal       bone    `json:"right_middle_distal,omit_empty"`
	RightRingProximal       bone    `json:"right_ring_proximal,omit_empty"`
	RightRingIntermediate   bone    `json:"right_ring_intermediate,omit_empty"`
	RightRingDistal         bone    `json:"right_ring_distal,omit_empty"`
	RightLittleProximal     bone    `json:"right_little_proximal,omit_empty"`
	RightLittleIntermediate bone    `json:"right_little_intermediate,omit_empty"`
	RightLittleDistal       bone    `json:"right_little_distal,omit_empty"`
	LastBone                bone    `json:"last_bone,omit_empty"`
}

func (vrm vrmType) say() {
	fmt.Println("joe")
}

// Listen for face data
func listenPerfectSync(address string, port int) {
	//listenRaw(address, port)
	listenOSC(address, port)
}

func genFloatSlice(interSlice []interface{}) ([]float32, error) {
	var floatSlice []float32

	for i := 0; i < len(interSlice); i++ {

		assertFloat, ok := (interSlice)[i].(float32)
		if !ok {
			return nil, fmt.Errorf("Could not assert to float")
		}

		floatSlice = append(floatSlice, assertFloat)
		fmt.Println(interSlice, assertFloat, floatSlice)

	}

	return floatSlice, nil

}

func parseOSC(msg *osc.Message) (string, []float32) {
	interKey := msg.Arguments[0]
	interValue := msg.Arguments[1:]

	var key string
	var value []float32

	key, _ = interKey.(string)

	for i, v := range interValue {

		value[i], _ = v.(float32)

	}

	return key, value
}

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

// Listen for face data through OSC from a device in the VMC protocol format
func listenOSC(address string, port int) {

	d := osc.NewStandardDispatcher()

	// Now to add whatever routes are needed, according to the protocol

	d.AddMsgHandler("/VMC/Ext/Blend/Val", func(msg *osc.Message) {

	})

	// Bone position and rotation request handler
	d.AddMsgHandler("/VMC/Ext/Bone/Pos", func(msg *osc.Message) {

		// First index of msg payload is the key name, all else is the actual data

		// Get key, convert to snake case
		key := fmt.Sprintf("%v", msg.Arguments[0])
		var snakeKey string
		if modifiedKey, err := camelToSnake(key); err != nil {
			return

		} else {
			snakeKey = modifiedKey
		}

		// Get msg args, type assert to float32
		var value []float32
		for _, v := range msg.Arguments[1:] {
			value = append(value, v.(float32))
		}

		// Assign rest of args to a new bone type
		newBone := bone{
			PositionX:   value[0],
			PositionY:   value[1],
			PositionZ:   value[2],
			QuaternionX: value[3],
			QuaternionY: value[4],
			QuaternionZ: value[5],
			QuaternionW: value[6],
		}

		// JSON byte representation of the new bone
		newBoneBytes, err := json.Marshal(newBone)
		if err != nil {
			return
		}

		// JSON byte representation of a new VRM with new bone for specified key
		newVrmBytes := []byte(
			fmt.Sprintf("{\"bones\":{\"%s\":%s}}", snakeKey, newBoneBytes),
		)

		// Finally, unmarshal the JSON representation of the VRM into liveVRM
		if err := json.Unmarshal(newVrmBytes, &liveVRM); err != nil {
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

func listenRaw(address string, port int) {

	// Valid address and port to listen on
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(address),
	}

	// Bind to address
	conn, err := net.ListenUDP("udp", &addr)
	defer conn.Close()
	if err != nil {
		log.Printf("Error listening on %s: %s\n", addr.String(), err)
		panic(err)
	}

	// UDP packet buffer
	var buf [12288]byte
	for {
		// Read from UDP connection raw bytes
		rlen, _, _ := conn.ReadFromUDP(buf[:])

		// Do stuff with it
		fmt.Println(rlen, string(buf[:]))
	}

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
func main() {

	go listenPerfectSync("0.0.0.0", 39540)

	router := mux.NewRouter()

	router.HandleFunc("/api/model", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received model API request from %s\n", r.RemoteAddr)
		ws, err := wsUpgrade(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		// Times per second to send the VMC data to client
		vmc_send_frequency := 60

		// Forever send to client VMC data
		for {

			ws.WriteJSON(liveVRM)
			time.Sleep(time.Duration(1e9 / vmc_send_frequency))

		}

	})

	http.ListenAndServe("127.0.0.1:3579", router)

}
