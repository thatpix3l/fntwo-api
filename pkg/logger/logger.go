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

package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L    = newLogger() // Normal logger
	S    = L.Sugar()   // Sugared logger
	atom = zap.NewAtomicLevel()
)

func SetLevel(level zapcore.Level) {
	atom.SetLevel(level)
}

func newLogger() *zap.Logger {

	// New encoder config for logger
	encoderConfig := zap.NewProductionEncoderConfig()

	// Add a few customizations for the logging config
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC1123Z)

	// Build the logger
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		atom,
	))

	return logger

}
