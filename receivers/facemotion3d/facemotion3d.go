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

package facemotion3d

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/obj"
	"github.com/thatpix3l/fntwo/receivers"
	"github.com/westphae/quaternion"
)

var (
	fm3dReceiver *receivers.MotionReceiver
)

// Extract one full frame of motion data. A valid full frame of data is basically anything between sets of the text: "___FACEMOTION3D".
func extractFrame(rawStr string) (string, error) {

	matchFrame := regexp.MustCompile("___FACEMOTION3D(.*)___FACEMOTION3D")
	frameMatches := matchFrame.FindStringSubmatch(rawStr)
	if len(frameMatches) == 0 {
		return "", errors.New("empty frame of data")
	}

	return strings.ReplaceAll(frameMatches[1], "___FACEMOTION3D", ""), nil

}

// Parse a full frame of motion data.
func parseFrame(frameStr string) {

	// All data is separated by the delimiter "|"
	payload := strings.Split(frameStr, "|")

	// For each data in the frame...
	for _, payloadStr := range payload {

		// If the current data is a blend shape (we know because it contains an "&" symbol)...
		if strings.Contains(payloadStr, "&") {

			// Skip Facemotion3D-specific blend shapes
			if strings.Contains(payloadStr, "FM_") {
				continue
			}

			// Skip empty keys
			if payloadStr == "" {
				continue
			}

			// The name and value are separated by a "&"
			singlePayload := strings.Split(payloadStr, "&")

			// The blend shape key is in camelCase, but we need it in PascalCase
			key := singlePayload[0]
			key = strings.ToUpper(key[0:1]) + key[1:]

			// Convert the blend shape value from a string to a float
			value, err := strconv.ParseFloat(singlePayload[1], 32)
			if err != nil {
				continue
			}
			// The blend shape values are in integer format from 0 to 100, but it has to be in decimal format from 0 to 1
			value = (value / 100)

			// Cast value to type BlendShape
			blendShape := obj.BlendShape(value)

			fm3dReceiver.VRM.WriteBlendShape(key, blendShape)

		}

		// If we're working with a bone (we know because it will contain an "#" symbol)...
		if strings.Contains(payloadStr, "#") {

			// The name and values are separated by a single "#"
			keyVal := strings.Split(payloadStr, "#")

			// Remove "=" char in key, convert from camelCase to PascalCase
			key := strings.ReplaceAll(keyVal[0], "=", "")
			key = strings.ToUpper(key[0:1]) + key[1:]

			// For each value for the current bone, convert it from a string to a float and store it in boneValues
			var boneValues []float32
			for _, v := range strings.Split(keyVal[1], ",") {

				rawFloat, err := strconv.ParseFloat(v, 32)
				if err != nil {
					log.Print(err)
					continue
				}

				boneValues = append(boneValues, float32(rawFloat))

			}

			// The bone rotations are in Euler. Instead, convert it to quaternion for the frontend

			// Divisor for certain bones to rotate normally when sent to the web client
			var divisor float32
			if key == "Head" {
				divisor = 32
			} else {
				divisor = 128
			}

			boneQuat := quaternion.FromEuler(
				float64(boneValues[0]/divisor),
				float64(boneValues[1]/divisor),
				-float64(boneValues[2]/divisor),
			)

			bone := obj.Bone{
				Rotation: obj.Rotation{
					Quaternion: obj.QuaternionRotation{
						X: float32(boneQuat.X),
						Y: float32(boneQuat.Y),
						Z: float32(boneQuat.Z),
						W: float32(boneQuat.W),
					},
				},
			}

			fm3dReceiver.VRM.WriteBone(key, bone)

		}

	}

}

// Tell a device with address to send the Facemotion3D data through TCP
func sendThroughTCP(address string) error {

	conn, err := net.Dial("udp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := fmt.Fprintf(conn, "StopStreaming_FACEMOTION3D"); err != nil {
		return err
	}
	time.Sleep(time.Second / 2)

	if _, err := fmt.Fprintf(conn, "FACEMOTION3D_OtherStreaming|protocol=tcp"); err != nil {
		return err
	}

	return nil

}

func listenTCP() {

	// Listen for new connections
	listener, err := net.Listen("tcp", fm3dReceiver.AppConfig.FM3DListen.String())
	if err != nil {
		log.Print(err)
		return
	}
	defer listener.Close()

	for {

		log.Printf("Telling device at \"%s\" to send motion Facemotion3D data through TCP", fm3dReceiver.AppConfig.FM3DDevice.IP())
		if err := sendThroughTCP(fm3dReceiver.AppConfig.FM3DDevice.IP() + ":49993"); err != nil {

			log.Print("Facemotion3D source error, waiting 3 seconds")
			time.Sleep(3 * time.Second)
			continue

		}

		// Accept new connection
		log.Print("Waiting for Facemotion3D client")
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			return
		}
		log.Print("Accepted new Facemotion3D client")

		var liveFrames string
		for {

			// Repeatedly read from connection new face data
			connBuf := make([]byte, 1024)
			_, err := conn.Read(connBuf)
			if err != nil {
				break
			}
			liveFrames += string(connBuf)

			// Attempt to pull the first valid frame of data.
			latestFrame, err := extractFrame(liveFrames)
			if err != nil {
				continue
			}

			// Parse the frame of data
			parseFrame(latestFrame)

			// Prune the frame of data that we just worked on, so we do not work with it on next iteration
			liveFrames = strings.ReplaceAll(liveFrames, latestFrame, "")

		}

		log.Print("Facemotion3D source disconnected, waiting 3 seconds")
		time.Sleep(3 * time.Second)

	}

}

// Create a new MotionReceiver.
// Uses the Facemotion3D app for face data. Internally, TCP is used to communicate with a device.
func New(appConfig *config.App) *receivers.MotionReceiver {

	fm3dReceiver = receivers.New(appConfig, listenTCP)
	return fm3dReceiver

}
