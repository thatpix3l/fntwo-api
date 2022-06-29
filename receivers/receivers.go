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

package receivers

import (
	"encoding/json"

	"github.com/thatpix3l/fntwo/obj"
)

// Update an existing VRM model with the given VRM data
func UpdateVRM(bonesMap map[string]obj.Bone, blendShapesMap map[string]float32, vrmPtr *obj.VRM) error {

	bonesMap["Hips"] = obj.Bone{
		Rotation: obj.QuaternionRotation{
			Y: bonesMap["Head"].Rotation.Y,
			W: 1,
		},
	}

	// Convert the blend shapes map into JSON bytes
	blendBuf, err := json.Marshal(blendShapesMap)
	if err != nil {
		return err
	}

	// Convert the bones map into JSON bytes
	boneBuf, err := json.Marshal(bonesMap)
	if err != nil {
		return err
	}

	// Unmarshal the blend shape bytes into the VRM
	if err := json.Unmarshal(blendBuf, &vrmPtr.BlendShapes.Face); err != nil {
		return err
	}

	// Unmarshal the bones bytes into the VRM
	if err := json.Unmarshal(boneBuf, &vrmPtr.Bones); err != nil {
		return err
	}

	return nil

}
