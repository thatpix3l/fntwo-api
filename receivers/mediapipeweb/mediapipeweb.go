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

		isClosed := false
		for !isClosed {

			// Mediapipe face mesh related data
			var mpFaceMesh mediapipeFacemesh
			if err := ws.ReadJSON(&mpFaceMesh); err != nil {
				log.Println(err)
				return
			}

			rotationRayVector := obj.Position{}
			rotationRayVectorCount := 0

			log.Println(mpFaceMesh)

			for _, triangleIndices := range faceMeshTriangleIndices {

				triangle := []obj.Position{
					mpFaceMesh.Landmarks[triangleIndices[0]],
					mpFaceMesh.Landmarks[triangleIndices[1]],
					mpFaceMesh.Landmarks[triangleIndices[2]],
				}

				// Normalize the world origin for all vertices of triangle, from Mediapipe's to threeJS
				triangle[0] = normalizePosition(triangle[0], mpWorldOrigin, mpFaceMesh.Video)
				triangle[1] = normalizePosition(triangle[1], mpWorldOrigin, mpFaceMesh.Video)
				triangle[2] = normalizePosition(triangle[2], mpWorldOrigin, mpFaceMesh.Video)

				// Create two vectors, each originating from different points and converging to the same point
				vector1 := directionVector(triangle[0], triangle[1])
				vector2 := directionVector(triangle[2], triangle[1])

				// Surface normal direction
				surfaceNormal := obj.Position{
					X: vector1.Y*vector2.Z - (vector1.Z * vector2.Y),
					Y: vector1.Z*vector2.X - (vector1.X * vector2.Z),
					Z: vector1.X*vector2.Y - (vector1.Y * vector2.X),
				}

				// Sum of direction for all triangle surface normals
				rotationRayVector.X += surfaceNormal.X
				rotationRayVector.Y += surfaceNormal.Y
				rotationRayVector.Z += surfaceNormal.Z

				rotationRayVectorCount++

			}

			// Divide the collective sum of all ray vectors by count of how many ray vectors there are.
			// Now you have the average of all ray vectors
			rotationRayVector.X /= float32(rotationRayVectorCount)
			rotationRayVector.Y /= float32(rotationRayVectorCount)
			rotationRayVector.Z /= float32(rotationRayVectorCount)

			dotProduct := (originRayVector.X * rotationRayVector.X) + (originRayVector.Y * rotationRayVector.Y) + (originRayVector.Z * rotationRayVector.Z)

			rotationRayVectorMagnitude := math.Pow(
				math.Pow(float64(rotationRayVector.X), 2)+
					math.Pow(float64(rotationRayVector.Y), 2)+
					math.Pow(float64(rotationRayVector.Z), 2),
				0.2,
			)

			log.Println(math.Acos(float64(dotProduct / (float32(originRayVectorMagnitude * rotationRayVectorMagnitude)))))

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
