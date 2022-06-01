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

// Spherical rotation
type SphericalRotation struct {
	AzimuthAngle float32 `json:"azimuth"`
	PolarAngle   float32 `json:"polar"`
}

// Transformational properties of the ThreeJS camera
type Camera struct {
	GazeTowards Position `json:"gaze_towards"`
	GazeFrom    Position `json:"gaze_from"`
}

// Properties that VRM models can have
type VRM struct {
	Bones       HumanBodyBones `json:"bones"`                  // All poseable bones, based off of Unity's HumanBodyBones
	BlendShapes BlendShapes    `json:"blend_shapes,omitempty"` // Updated blend shape data
}

// All available VRM blend shapes
type BlendShapes struct {
	Face FaceBlendShapes `json:"face,omitempty"`
}

// The available face blend shapes to modify, based off of Apple's 52 BlendShape AR-kit spec
// Some VRM models can miss a few, or even all blend shapes, so it is okay to not send any data if missing
type FaceBlendShapes struct {
	EyeBlinkLeft        float32 `json:"EyeBlinkLeft,omitempty"`
	EyeLookDownLeft     float32 `json:"EyeLookDownLeft,omitempty"`
	EyeLookInLeft       float32 `json:"EyeLookInLeft,omitempty"`
	EyeLookOutLeft      float32 `json:"EyeLookOutLeft,omitempty"`
	EyeLookUpLeft       float32 `json:"EyeLookUpLeft,omitempty"`
	EyeSquintLeft       float32 `json:"EyeSquintLeft,omitempty"`
	EyeWideLeft         float32 `json:"EyeWideLeft,omitempty"`
	EyeBlinkRight       float32 `json:"EyeBlinkRight,omitempty"`
	EyeLookDownRight    float32 `json:"EyeLookDownRight,omitempty"`
	EyeLookInRight      float32 `json:"EyeLookInRight,omitempty"`
	EyeLookOutRight     float32 `json:"EyeLookOutRight,omitempty"`
	EyeLookUpRight      float32 `json:"EyeLookUpRight,omitempty"`
	EyeSquintRight      float32 `json:"EyeSquintRight,omitempty"`
	EyeWideRight        float32 `json:"EyeWideRight,omitempty"`
	JawForward          float32 `json:"JawForward,omitempty"`
	JawLeft             float32 `json:"JawLeft,omitempty"`
	JawRight            float32 `json:"JawRight,omitempty"`
	JawOpen             float32 `json:"JawOpen,omitempty"`
	MouthClose          float32 `json:"MouthClose,omitempty"`
	MouthFunnel         float32 `json:"MouthFunnel,omitempty"`
	MouthPucker         float32 `json:"MouthPucker,omitempty"`
	MouthLeft           float32 `json:"MouthLeft,omitempty"`
	MouthRight          float32 `json:"MouthRight,omitempty"`
	MouthSmileLeft      float32 `json:"MouthSmileLeft,omitempty"`
	MouthSmileRight     float32 `json:"MouthSmileRight,omitempty"`
	MouthFrownLeft      float32 `json:"MouthFrownLeft,omitempty"`
	MouthFrownRight     float32 `json:"MouthFrownRight,omitempty"`
	MouthDimpleLeft     float32 `json:"MouthDimpleLeft,omitempty"`
	MouthDimpleRight    float32 `json:"MouthDimpleRight,omitempty"`
	MouthStretchLeft    float32 `json:"MouthStretchLeft,omitempty"`
	MouthStretchRight   float32 `json:"MouthStretchRight,omitempty"`
	MouthRollLower      float32 `json:"MouthRollLower,omitempty"`
	MouthRollUpper      float32 `json:"MouthRollUpper,omitempty"`
	MouthShrugLower     float32 `json:"MouthShrugLower,omitempty"`
	MouthShrugUpper     float32 `json:"MouthShrugUpper,omitempty"`
	MouthPressLeft      float32 `json:"MouthPressLeft,omitempty"`
	MouthPressRight     float32 `json:"MouthPressRight,omitempty"`
	MouthLowerDownLeft  float32 `json:"MouthLowerDownLeft,omitempty"`
	MouthLowerDownRight float32 `json:"MouthLowerDownRight,omitempty"`
	MouthUpperUpLeft    float32 `json:"MouthUpperUpLeft,omitempty"`
	MouthUpperUpRight   float32 `json:"MouthUpperUpRight,omitempty"`
	BrowDownLeft        float32 `json:"BrowDownLeft,omitempty"`
	BrowDownRight       float32 `json:"BrowDownRight,omitempty"`
	BrowInnerUp         float32 `json:"BrowInnerUp,omitempty"`
	BrowOuterUpLeft     float32 `json:"BrowOuterUpLeft,omitempty"`
	BrowOuterUpRight    float32 `json:"BrowOuterUpRight,omitempty"`
	CheekPuff           float32 `json:"CheekPuff,omitempty"`
	CheekSquintLeft     float32 `json:"CheekSquintLeft,omitempty"`
	CheekSquintRight    float32 `json:"CheekSquintRight,omitempty"`
	NoseSneerLeft       float32 `json:"NoseSneerLeft,omitempty"`
	NoseSneerRight      float32 `json:"NoseSneerRight,omitempty"`
	TongueOut           float32 `json:"TongueOut,omitempty"`
}

// Properties of a single VRM Bone
type Bone struct {
	Position Position           `json:"position"`
	Rotation QuaternionRotation `json:"rotation"`
}

// All bones used in a VRM model, based off of Unity's HumanBodyBones
type HumanBodyBones struct {
	Hips                    Bone `json:"Hips"`
	LeftUpperLeg            Bone `json:"LeftUpperLeg"`
	RightUpperLeg           Bone `json:"RightUpperLeg"`
	LeftLowerLeg            Bone `json:"LeftLowerLeg"`
	RightLowerLeg           Bone `json:"RightLowerLeg"`
	LeftFoot                Bone `json:"LeftFoot"`
	RightFoot               Bone `json:"RightFoot"`
	Spine                   Bone `json:"Spine"`
	Chest                   Bone `json:"Chest"`
	UpperChest              Bone `json:"UpperChest"`
	Neck                    Bone `json:"Neck"`
	Head                    Bone `json:"Head"`
	LeftShoulder            Bone `json:"LeftShoulder"`
	RightShoulder           Bone `json:"RightShoulder"`
	LeftUpperArm            Bone `json:"LeftUpperArm"`
	RightUpperArm           Bone `json:"RightUpperArm"`
	LeftLowerArm            Bone `json:"LeftLowerArm"`
	RightLowerArm           Bone `json:"RightLowerArm"`
	LeftHand                Bone `json:"LeftHand"`
	RightHand               Bone `json:"RightHand"`
	LeftToes                Bone `json:"LeftToes"`
	RightToes               Bone `json:"RightToes"`
	LeftEye                 Bone `json:"LeftEye"`
	RightEye                Bone `json:"RightEye"`
	Jaw                     Bone `json:"Jaw"`
	LeftThumbProximal       Bone `json:"LeftThumbProximal"`
	LeftThumbIntermediate   Bone `json:"LeftThumbIntermediate"`
	LeftThumbDistal         Bone `json:"LeftThumbDistal"`
	LeftIndexProximal       Bone `json:"LeftIndexProximal"`
	LeftIndexIntermediate   Bone `json:"LeftIndexIntermediate"`
	LeftIndexDistal         Bone `json:"LeftIndexDistal"`
	LeftMiddleProximal      Bone `json:"LeftMiddleProximal"`
	LeftMiddleIntermediate  Bone `json:"LeftMiddleIntermediate"`
	LeftMiddleDistal        Bone `json:"LeftMiddleDistal"`
	LeftRingProximal        Bone `json:"LeftRingProximal"`
	LeftRingIntermediate    Bone `json:"LeftRingIntermediate"`
	LeftRingDistal          Bone `json:"LeftRingDistal"`
	LeftLittleProximal      Bone `json:"LeftLittleProximal"`
	LeftLittleIntermediate  Bone `json:"LeftLittleIntermediate"`
	LeftLittleDistal        Bone `json:"LeftLittleDistal"`
	RightThumbProximal      Bone `json:"RightThumbProximal"`
	RightThumbIntermediate  Bone `json:"RightThumbIntermediate"`
	RightThumbDistal        Bone `json:"RightThumbDistal"`
	RightIndexProximal      Bone `json:"RightIndexProximal"`
	RightIndexIntermediate  Bone `json:"RightIndexIntermediate"`
	RightIndexDistal        Bone `json:"RightIndexDistal"`
	RightMiddleProximal     Bone `json:"RightMiddleProximal"`
	RightMiddleIntermediate Bone `json:"RightMiddleIntermediate"`
	RightMiddleDistal       Bone `json:"RightMiddleDistal"`
	RightRingProximal       Bone `json:"RightRingProximal"`
	RightRingIntermediate   Bone `json:"RightRingIntermediate"`
	RightRingDistal         Bone `json:"RightRingDistal"`
	RightLittleProximal     Bone `json:"RightLittleProximal"`
	RightLittleIntermediate Bone `json:"RightLittleIntermediate"`
	RightLittleDistal       Bone `json:"RightLittleDistal"`
	LastBone                Bone `json:"LastBone"`
}
