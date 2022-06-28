package virtualmotioncapture

import (
	"fmt"
	"log"

	"github.com/hypebeast/go-osc/osc"
	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/obj"
	"github.com/thatpix3l/fntwo/receivers"
)

var (
	vmcReceiver config.MotionReceiver
)

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

// Start listening for VMC messages to modify the VRM data
func listenVMC() {

	// Listen for face and bone data through OSC from a device in the VMC protocol format
	log.Printf("Listening for VMC model transformation data on %s", vmcReceiver.AppCfg.GetVmcServerAddress())

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
		blendShapesMap := make(map[string]float32)
		blendShapesMap[key] = blendValue

		// Update VRM with new blend shape data
		receivers.UpdateVRM(nil, blendShapesMap, vmcReceiver.VRM)

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
		bonesMap := make(map[string]obj.Bone)

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

		// The bone with the name from the OSC message will have this new bone data
		bonesMap[key] = newBone

		// Update VRM with new blend shape data
		receivers.UpdateVRM(bonesMap, nil, vmcReceiver.VRM)

	})

	// OSC server configuration
	addr := vmcReceiver.AppCfg.GetVmcServerAddress()
	server := &osc.Server{
		Addr:       addr,
		Dispatcher: d,
	}

	// Blocking listen and serve
	server.ListenAndServe()

}

// Create a new VirtualMotionCapture receiver.
// Internally, uses OSC messaging, which in turn uses UDP for low-latency motion parsing
func New(vrmPtr *obj.VRM, appCfgPtr *config.App) config.MotionReceiver {

	vmcReceiver = config.MotionReceiver{
		VRM:    vrmPtr,
		AppCfg: appCfgPtr,
		Start:  listenVMC,
	}

	return vmcReceiver

}
