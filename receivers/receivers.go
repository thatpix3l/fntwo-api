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
	"github.com/thatpix3l/fntwo/config"
	"github.com/thatpix3l/fntwo/obj"
)

type MotionReceiver struct {
	AppConfig     *config.App // Pointer an existing app config, for reading various settings.
	VRM           obj.VRM     // VRM object to transform in 3D space.
	startCallback func()      // Callback for starting the receiver
	stopCallback  func()      // Callback for stopping the receiver
}

// Create a new motion receiver.
func New(appConfig *config.App, start func(), stop func()) *MotionReceiver {

	return &MotionReceiver{
		AppConfig:     appConfig,
		VRM:           obj.NewVRM(),
		startCallback: start,
		stopCallback:  stop,
	}

}

// Start motion receiver in background
func (m *MotionReceiver) Start() *MotionReceiver {
	go m.startCallback()
	return m
}

// Stop motion receiver
func (m *MotionReceiver) Stop() *MotionReceiver {
	m.stopCallback()
	return m
}
