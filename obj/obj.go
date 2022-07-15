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

// 3D-related structures for different objects used throughout both the frontend and backend
package obj

import (
	"sync"
)

// Positioning
type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// Quaternion rotation
type QuaternionRotation struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
	W float32 `json:"w"`
}

// Properties of a single VRM Bone
type Bone struct {
	Position Position           `json:"position"`
	Rotation QuaternionRotation `json:"rotation"`
}

// Transformational properties of the ThreeJS camera
type Camera struct {
	GazeTowards Position `json:"gaze_towards"`
	GazeFrom    Position `json:"gaze_from"`
}

// Primitive VRM blend shape value. By default, a float32
type BlendShape float32

type BlendShapes map[string]BlendShape

type Bones map[string]Bone

// VRM model for 3D-transformation purposes
type VRM struct {
	Bones            Bones       `json:"bones"`        // All poseable bones, based off of Unity's HumanBodyBones
	BlendShapes      BlendShapes `json:"blend_shapes"` // All blend shapes, unique to VRM model
	bonesMutex       *sync.RWMutex
	blendShapesMutex *sync.RWMutex
	readCallback     func(vrm *VRM)
}

// Create a new VRM object
func NewVRM() *VRM {

	bones := make(Bones)
	blendShapes := make(BlendShapes)

	return &VRM{
		Bones:            bones,
		BlendShapes:      blendShapes,
		bonesMutex:       &sync.RWMutex{},
		blendShapesMutex: &sync.RWMutex{},
	}

}

// Run a function to safely read VRM data
func (v *VRM) Read(callback func(vrm *VRM)) {

	// Lock VRM for safe reading
	v.bonesMutex.RLock()
	v.blendShapesMutex.RLock()
	defer v.bonesMutex.RUnlock()
	defer v.blendShapesMutex.RUnlock()

	// Process VRM data
	callback(v)

}

func (v *VRM) WriteBone(key string, value Bone) {

	// Lock VRM for safe writing
	v.bonesMutex.Lock()
	defer v.bonesMutex.Unlock()

	// Modify VRM bones
	v.Bones[key] = value

}

func (v *VRM) WriteBlendShape(key string, value BlendShape) {

	// Lock VRM for safe writing
	v.blendShapesMutex.Lock()
	defer v.blendShapesMutex.Unlock()

	// Modify VRM blend shapes
	v.BlendShapes[key] = value

}
