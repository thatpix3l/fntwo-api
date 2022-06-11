package facemotion3d

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/obj"
)

var (
	fm3dReceiver config.MotionReceiver
)

func extractFrame(rawStr string) string {

	matchFrame := regexp.MustCompile("___FACEMOTION3D(.*)___FACEMOTION3D")
	frameMatches := matchFrame.FindStringSubmatch(rawStr)
	if len(frameMatches) == 0 {
		return ""
	}

	return strings.ReplaceAll(frameMatches[1], "___FACEMOTION3D", "")

}

func parseFrame(frameStr string) (map[string]float32, map[string]obj.Bone) {

	// Face and bone blend shape keys with associated float values
	newBlendShapes := make(map[string]float32)
	newBones := make(map[string]obj.Bone)

	payload := strings.Split(frameStr, "|")
	for _, val := range payload {

		// If the type payload is a blend shape...
		if strings.Contains(val, "&") {

			// Skip Facemotion3D-specific blend shapes
			if strings.Contains(val, "FM_") {
				continue
			}

			// Skip empty keys
			if val == "" {
				continue
			}

			// The name and value are separated by a "&"
			singlePayload := strings.Split(val, "&")

			// The blend shape keys are in camelCase, but we need them in PascalCase
			rawKey := singlePayload[0]
			blendKey := strings.ToUpper(rawKey[0:1]) + rawKey[1:]

			// Convert the blend shape value from a string to a float
			rawValue, err := strconv.ParseFloat(singlePayload[1], 32)
			if err != nil {
				continue
			}

			// The blend shape values are in integer format from 0 to 100, but it has to be in decimal format from 0 to 1
			blendValue := (rawValue / 100)

			newBlendShapes[blendKey] = float32(blendValue)

		}

		// If we're working with a bone
		if strings.Contains(val, "#") {

			// The name and values are separated by a single "#"
			keyVal := strings.Split(val, "#")

			// Remove "=" char in key, convert from camelCase to PascalCase
			boneKey := strings.ReplaceAll(keyVal[0], "=", "")
			boneKey = strings.ToUpper(boneKey[0:1]) + boneKey[1:]

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

			newBones[boneKey] = obj.Bone{
				Rotation: obj.QuaternionRotation{
					X: boneValues[0],
					Y: boneValues[1],
					Z: boneValues[2],
					W: 1,
				},
			}

		}

	}

	return newBlendShapes, newBones

}

func listenTCP() {

	// Tell a device to send the Facemotion3D data through TCP
	log.Print("Telling phone to send motion data through TCP")
	motionSrc, err := net.Dial("udp", "10.0.1.220:49993")
	if err != nil {
		log.Print(err)
		return
	}
	defer motionSrc.Close()
	fmt.Fprintf(motionSrc, "StopStreaming_FACEMOTION3D")
	fmt.Fprintf(motionSrc, "FACEMOTION3D_OtherStreaming|protocol=tcp")

	// Listen for new connections
	listener, err := net.Listen("tcp", ":49986")
	if err != nil {
		log.Print(err)
		return
	}
	defer listener.Close()

	var liveFrames string

	for {

		// Accept new connection
		log.Print("Waiting for Facemotion3D client")
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			return
		}
		defer conn.Close()
		log.Print("Accepted new Facemotion3D client")

		for {

			// Repeatedly read from connection new face data
			connBuf := make([]byte, 1024)
			_, err := conn.Read(connBuf)
			if err != nil {
				break
			}
			liveFrames += string(connBuf)

			// Attempt to pull the first valid frame of data.
			latestFrame := extractFrame(liveFrames)
			if latestFrame == "" {
				continue
			}

			// Parse the frame of data
			blendShapesMap, bonesMap := parseFrame(latestFrame)

			// Convert the blend shapes map into JSON bytes
			blendBuf, err := json.Marshal(blendShapesMap)
			if err != nil {
				log.Print(err)
				continue
			}

			// Convert the bones map into JSON bytes
			boneBuf, err := json.Marshal(bonesMap)
			if err != nil {
				log.Print(err)
				continue
			}

			// Unmarshal the blend shape bytes into the VRM
			if err := json.Unmarshal(blendBuf, &fm3dReceiver.VRM.BlendShapes.Face); err != nil {
				log.Print(err)
				continue
			}

			// Unmarshal the bones bytes into the VRM
			if err := json.Unmarshal(boneBuf, &fm3dReceiver.VRM.Bones); err != nil {
				log.Print(err)
				continue
			}

			// Prune the frame of data that we just worked on, so we do not work with it on next iteration
			liveFrames = strings.ReplaceAll(liveFrames, latestFrame, "")

		}

	}

}

// Create a new Facemotion3D receiver.
// Uses the Facemotion3D app for face data. Internally, TCP is used to communicate with a device.
func New(vrmPtr *obj.VRM, appCfgPtr *config.App) config.MotionReceiver {

	fm3dReceiver = config.MotionReceiver{
		VRM:    vrmPtr,
		AppCfg: appCfgPtr,
		Start:  listenTCP,
	}

	return fm3dReceiver

}
