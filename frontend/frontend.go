package frontend

import (
	"embed"
	"io/fs"
)

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
