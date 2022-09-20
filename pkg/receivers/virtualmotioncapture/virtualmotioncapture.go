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

package virtualmotioncapture

import (
	"fmt"
	"log"
	"unicode"

	"github.com/hypebeast/go-osc/osc"
	"github.com/thatpix3l/fntwo/pkg/config"
	"github.com/thatpix3l/fntwo/pkg/obj"
	"github.com/thatpix3l/fntwo/pkg/receivers"
)

var (
	vmcReceiver *receivers.MotionReceiver
	server      *osc.Server
)

// Assuming everything after the first index is bone data, type assert it as a slice of float32
// The positioning of the data is special, where the index is as follows:
// index 0, 1, 2: bone position X, Y, Z
// index 3, 4, 5, 6: bone quaternion rotation X, Y, Z, W
func parseBone(msg *osc.Message) ([]float64, error) {

	// Slice of bone data parameters
	var boneData []float64

	// For each OSC message index, skipping the first index...
	for _, v := range msg.Arguments[1:] {
		coord, ok := v.(float64)
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
	log.Println(vmcReceiver.AppConfig.APIListen)
	log.Printf("Listening for VMC model transformation data on %s", vmcReceiver.AppConfig.VMCListen)

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
		value, ok := msg.Arguments[1].(float32)
		if !ok {
			return
		}

		// Set max and min for blendValue to between 0 and 1
		if value > 1 {
			value = 1
		}

		if value < 0 {
			value = 0
		}

		vmcReceiver.VRM.WriteBlendShape(key, obj.BlendShape(value))

	})

	// Bone position and rotation request handler
	d.AddMsgHandler("/VMC/Ext/Bone/Pos", func(msg *osc.Message) {

		// Bone name
		key, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}

		// Make the key's first letter upper case
		key = string(unicode.ToUpper([]rune(key)[0])) + key[1:]

		// Bone transformation parameters slice
		value, err := parseBone(msg)
		if err != nil {
			return
		}

		// New bone structure
		bone := obj.Bone{
			Position: obj.Position{
				X: value[0],
				Y: value[1],
				Z: value[2],
			},
			Rotation: obj.Rotation{
				Quaternion: obj.QuaternionRotation{
					X: value[3],
					Y: value[4],
					Z: value[5],
					W: value[6],
				},
			},
		}

		// Attach bone to the receiver's referenced bone map
		vmcReceiver.VRM.WriteBone(key, bone)

	})

	// OSC server configuration
	server = &osc.Server{
		Addr:       vmcReceiver.AppConfig.VMCListen.String(),
		Dispatcher: d,
	}

	// Blocking listen and serve
	server.ListenAndServe()

}

// Create a new MotionReceiver.
// Uses the VMC protocol, a subset of the OSC protocol, which internally uses UDP for low-latency motion parsing.
func New(appConfig *config.App) *receivers.MotionReceiver {

	vmcReceiver = receivers.New(appConfig, listenVMC)
	return vmcReceiver

}
