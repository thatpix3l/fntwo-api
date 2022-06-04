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

package frontend

import (
	"embed"
	"io/fs"
)

//go:generate echo "Building frontend..."
//go:generate npm --prefix static run build

var (

	//go:embed static/build
	staticFS embed.FS
)

func FS() (fs.FS, error) {
	rootFS, err := fs.Sub(staticFS, "static/build")
	if err != nil {
		return nil, err
	}

	return rootFS, nil

}
