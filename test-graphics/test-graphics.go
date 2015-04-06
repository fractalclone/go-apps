// Copyright (c) 2015 Jean-Paul Etienne <fractalclone@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// xres and yres constants can be adjusted to change
// the size of the X11 window
// The num_rectangles constant defines the number of
// rectangles to be displayed in the X11 window.
// (tested up to 2 million rectangles)
package main

import (
	"code.google.com/p/x-go-binding/xgb"
	"fmt"
	"time"
	"math/rand"
)

type Rect struct {
	x		int
	prevx		int
	y		int
	prevy		int
	width		int
	height		int
	movx		int
	movy		int
	speed		time.Duration
	render_done	chan bool
}

type Graphic struct {
	c		*xgb.Conn
	win		xgb.Id
	bg		xgb.Id
	fg		xgb.Id
	xres		int
	yres		int
}

const xres int = 2048
const yres int = 1400
const num_rectangles int = 1000

func randInt(min int, max int) int {
	return min + rand.Intn(max - min)
}

func generate_rec(g *Graphic, c chan *Rect) {
	rec := new(Rect)
	rec.width = randInt(20, 50)
	rec.height = rec.width
	rec.prevx = -1
	rec.prevy = -1
	rec.x = randInt(0, g.xres - rec.width)
	rec.y = randInt(0, g.yres - rec.height)
	rec.speed = time.Duration(randInt(10, 50))
	rec.movx = randInt(0, 10) - 5
	if (rec.movx == 0) {
		rec.movx = 1
	}
	rec.movy = randInt(0, 10) - 5
	if (rec.movy == 0) {
		rec.movy = -1
	}
	rec.render_done = make(chan bool)
	go move_rec(g, rec, c);
}

func move_rec(g* Graphic, r *Rect, c chan *Rect) {
	for (true) {
		// send Rect element to renderer
		c <- r

		// wait for renderer to complete Rect rendering
		<- r.render_done

		time.Sleep(r.speed * time.Millisecond)

		// move rectangle
		r.prevx = r.x
		r.prevy = r.y

		if ((r.x + r.movx + r.width) > g.xres ||
			(r.x + r.movx) <= 0) {
			r.movx = -r.movx
		}

		if ((r.y + r.movy + r.height) > g.yres ||
			(r.y + r.movy) <= 0) {
			r.movy = -r.movy
		}

		r.x += r.movx
		r.y += r.movy
	}
}

func renderer(g *Graphic, c chan *Rect) {
	var rec *Rect;
	rectangles := []xgb.Rectangle{{0, 0, 0, 0}}

	for (true) {
		// wait for Rect element render on screen
		rec = <- c

		rectangles[0].Width = uint16(rec.width)
		rectangles[0].Height = uint16(rec.height)

		// if previous x and y coordinates are >= 0
		// erase rectangle element at previous x and y coordinates
		if (rec.prevx >= 0 && rec.prevy >= 0) {
			rectangles[0].X = int16(rec.prevx)
			rectangles[0].Y = int16(rec.prevy)
			g.c.PolyRectangle(g.win, g.bg, rectangles)
		}

		// display rectangle element at new x and y coordinates
		rectangles[0].X = int16(rec.x)
		rectangles[0].Y = int16(rec.y)
		g.c.PolyRectangle(g.win, g.fg, rectangles)

		// inform about rendering completion for Rect element
		rec.render_done <- true
	}
}

func init_X11() (*Graphic, error) {
	var g = new(Graphic)
	var err error

	g.c, err = xgb.Dial("")
	if (err != nil) {
		fmt.Println(err)
		return nil, err
	}

	g.xres = xres
	g.yres = yres

	s := g.c.DefaultScreen()
	g.win = s.Root

	// create black foreground graphic context
	g.fg = g.c.NewId()
	mask := uint32(xgb.GCForeground | xgb.GCGraphicsExposures)
	values := []uint32{s.BlackPixel, 0}
	g.c.CreateGC(g.fg, g.win, mask, values)

	// create white foreground graphic context to erase
	// rectangles while moving them
	g.bg = g.c.NewId()
	mask = uint32(xgb.GCForeground | xgb.GCGraphicsExposures)
	values[0] = s.WhitePixel
	g.c.CreateGC(g.bg, g.win, mask, values)

	// create white background graphic context
	bg := g.c.NewId()
	mask = uint32(xgb.GCBackground | xgb.GCGraphicsExposures)
	values[0] = s.WhitePixel
	g.c.CreateGC(bg, g.win, mask, values)

	// create X11 window
	g.win = g.c.NewId()
	mask = xgb.CWBackPixel | xgb.CWEventMask
	values[1] = xgb.EventMaskExposure | xgb.EventMaskKeyPress
	g.c.CreateWindow(0,
		g.win,
		s.Root,
		0,
		0,
		uint16(xres),
		uint16(yres),
		10,
		xgb.WindowClassInputOutput,
		s.RootVisual,
		mask,
		values)
	g.c.MapWindow(g.win)
	return g, nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	g, err := init_X11();
	if (err != nil) {
		return
	}

	defer g.c.Close()

	// create buffered channel of num_rectangles
	// to synchronize renderer with mov_rec routines
	c_rec := make(chan *Rect, num_rectangles)

	// start renderer
	go renderer(g, c_rec);

	// generate routines for moving rectangles
	for i := 0; i < num_rectangles; i++ {
		go generate_rec(g, c_rec)
	}

	for {
		event, err := g.c.WaitForEvent()
		if err != nil {
			fmt.Println(err)
			return
		}

		switch event.(type) {
		case xgb.KeyPressEvent:
			return
		}
	}
}
