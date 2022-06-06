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
	tracker config.MotionReceiver
)

func extractFrame(rawStr string) string {

	matchSingleFrame := regexp.MustCompile("___FACEMOTION3D(.*)__FACEMOTION3D")
	singleFrame := matchSingleFrame.FindString(rawStr)
	return singleFrame

}

func parseFrame(frameStr string) map[string]float32 {

	// Face blend shape keys with associated float values
	newBlendShapes := make(map[string]float32)

	payload := strings.Split(frameStr, "|")
	for _, val := range payload {

		// If the type payload is a blend shape...
		if strings.Contains(val, "&") {

			// Skip Facemotion3D-specific blend shapes
			if strings.Contains(val, "FM_") {
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

	}

	return newBlendShapes

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

			// Parse the frame of data into JSON bytes
			blendBuf, err := json.Marshal(parseFrame(latestFrame))
			if err != nil {
				log.Print(err)
				continue
			}

			// Unmarshal the data into the VRM face
			if err := json.Unmarshal(blendBuf, &tracker.VRM.BlendShapes.Face); err != nil {
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

	tracker = config.MotionReceiver{
		VRM:    vrmPtr,
		AppCfg: appCfgPtr,
		Start:  listenTCP,
	}

	return tracker

}
