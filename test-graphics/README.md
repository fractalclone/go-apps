# test-graphics

**test-graphics** is a simple X11-based application, which generates moving rectangles of random size, (x,y) coordinates, speed and direction in an X11 window. It is used to test goroutines in terms of performance, scalability and memory consumption.

Each generated rectangle is handled via a `move_rec` routine, which sends periodically to the `renderer` routine the rectangle structure to display. Synchronization between `move_rec` and `renderer` routines is handled via:

 - a buffered channel, which is used by the `move_rec` routines to send pointers of rectangle elements to the `renderer` routine
 - a simple boolean channel, which is used by the `renderer` routine to inform to a given `move_rec` routine when the rendering of a rectangle element has completed

The `xres` and `yres` constants can be adjusted to change the size of the X11 window.
The `num_rectangles` constant defines the number of rectangles to be displayed in the X11 window.

### Installation
**test-graphics** requires the use of `x-go-binding/xgb` package. This package can be installed as follows:
```sh
$ cd path_to_install_my_go_package
$ go get code.google.com/p/x-go-binding/xgb
```
### Run
```sh
$ export GOPATH=$GOPATH:path_to_install_my_go_package
$ go run ./test-graphics.go
```
