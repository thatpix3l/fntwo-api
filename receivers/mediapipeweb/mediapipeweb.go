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

package mediapipeweb

import (
	"log"
	"math"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/helper"
	"github.com/thatpix3l/fntwo/obj"
	"github.com/thatpix3l/fntwo/receivers"
)

type videoMetadata struct {
	Width  int `json:"width"`  // Width of the source video
	Height int `json:"height"` // Height of the source video
}

type mediapipeFacemesh struct {
	Landmarks []obj.Position `json:"landmarks"` // List of landmark positions
	Video     videoMetadata  `json:"video"`     // Metadata of source video used in Mediapipe
}

var (
	mpReceiver    *receivers.MotionReceiver
	mpWorldOrigin = obj.Position{
		X: 0.5,
		Y: 0.5,
		Z: 0,
	}

	originRayVector = obj.Position{
		X: 0,
		Y: 0,
		Z: 1,
	}

	originRayVectorMagnitude = math.Pow(
		math.Pow(float64(originRayVector.X), 2)+
			math.Pow(float64(originRayVector.Y), 2)+
			math.Pow(float64(originRayVector.Z), 2),
		0.2,
	)
)

const (
	coordMultiplier = 2
)

func normalizePosition(position obj.Position, worldOrigin obj.Position, video videoMetadata) obj.Position {
	return obj.Position{
		X: (position.X - worldOrigin.X) * float32(video.Width),
		Y: (position.Y - worldOrigin.Y) * float32(video.Height),
		Z: (position.Z - worldOrigin.Z) * float32(coordMultiplier),
	}
}

func centroid(positions ...obj.Position) obj.Position {

	sum := obj.Position{
		X: 0,
		Y: 0,
		Z: 0,
	}
	for _, position := range positions {
		sum.X += position.X
		sum.Y += position.Y
		sum.Z += position.Z
	}

	sum.X /= float32(len(positions))
	sum.Y /= float32(len(positions))
	sum.Z /= float32(len(positions))

	return sum
}

func directionVector(from obj.Position, to obj.Position) obj.Position {
	return obj.Position{
		X: to.X - from.X,
		Y: to.Y - from.Y,
		Z: to.Z - from.Y,
	}
}

// Start
func listenMediapipeWeb() {

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		ws, err := helper.WebSocketUpgrade(w, r)
		if err != nil {
			log.Println(err)
			return
		}

		for {

			// Mediapipe face mesh related data
			var mpFaceMesh []obj.Position
			if err := ws.ReadJSON(&mpFaceMesh); err != nil {
				log.Println(err)
				return
			}

		}

	})

	http.ListenAndServe("0.0.0.0:2332", router)

}

// Create and return reference to a MotionReceiver.
// Listens for WebSocket connections
func New(appConfig *config.App) *receivers.MotionReceiver {

	mpReceiver = receivers.New(appConfig, listenMediapipeWeb)
	return mpReceiver

}
