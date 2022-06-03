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

// This is technically a different program altogether, just to pull the version information from the frontend and backend
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func cmdOutput(cmdName string, args ...string) (string, error) {

	cmdOutput, err := exec.Command(cmdName, args...).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(cmdOutput), "\n"), nil

}

func main() {

	// Store frontend's git version
	frontendVersion, err := cmdOutput("git", "--git-dir=frontend/static/.git", "--work-tree=frontend/static", "describe", "--tags")
	if err != nil {
		log.Fatal(err)
	}

	//Store backend's git version
	backendVersion, err := cmdOutput("git", "describe", "--tags")

	// Write to new version.txt
	versionFile, err := os.Create("./version/version.txt")
	if err != nil {
		log.Print(err)
		return
	}
	defer versionFile.Close()

	versionFile.WriteString(fmt.Sprintf("(WebUI: %s, Backend API: %s)", frontendVersion, backendVersion))

}
