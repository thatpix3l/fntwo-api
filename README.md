# Function Two
A project comprised of one, possibly two, *maybe even three* functions to make VTubing easy.
While existing VTube software for 3D avatars are available, none are as truly cross-platform as this one.

## Acquire this software...

### ...by downloading.
Yeah, no. Compile and run from source, or wait until this message changes :D

### ...by compiling.
There's actually two things needed to compile: the frontend & backend.

The frontend is an NPM project, so install Node and NPM, or whatever alternative there is for NPM.
Then, for the frontend, do:
```sh
git clone github.com/thatpix3l/fntwo-frontend
cd fntwo-frontend
npm install
npm run dev
```

As for the backend:
```sh
git clone github.com/thatpix3l/fntwo
cd fntwo
go mod download
go run .
```

## After starting...
...open both your desktop browser and the OBS browser source to `127.0.0.1:10001`.
In the desktop browser, pressing ESCAPE opens the main menu.
In the main menu, dragging and dropping a `.vrm` file onto the left pane loads the new model.
Controls are as follows:
- left-click and drag **rotates the camera**
- right-click and drag **pans the camera** up & down, left & right.
- W, A, S, D **pans the camera** forward & backward, left & right.