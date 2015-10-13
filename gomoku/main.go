// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

// An app that demonstrates the sprite package.
//
// Note: This demo is an early preview of Go 1.5. In order to build this
// program as an Android APK using the gomobile tool.
//
// See http://godoc.org/golang.org/x/mobile/cmd/gomobile to install gomobile.
//
// Get the sprite example and use gomobile to build or install it on your device.
//
//   $ go get -d golang.org/x/mobile/example/sprite
//   $ gomobile build golang.org/x/mobile/example/sprite # will build an APK
//
//   # plug your Android device to your computer or start an Android emulator.
//   # if you have adb installed on your machine, use gomobile install to
//   # build and deploy the APK to an Android target.
//   $ gomobile install golang.org/x/mobile/example/sprite
//
// Switch to your device or emulator to start the Basic application from
// the launcher.
// You can also run the application on your desktop by running the command
// below. (Note: It currently doesn't work on Windows.)
//   $ go install golang.org/x/mobile/example/sprite && sprite
package main

import (
	"bytes"
	"image"
	"log"
	"time"

	_ "image/png"

	"./board"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

const (
	SPACE = 0
	BLACK = 1
	WHITE = 2
)

var (
	startTime = time.Now()
	images    *glutil.Images
	eng       sprite.Engine
	scene     *sprite.Node
	fps       *debug.FPS

	endFlag   bool
	goisiTexs []sprite.SubTex
	loadscene bool
	whichTurn int
	b         *board.Board

	prevPosX int
	prevPosY int
	prevN    *sprite.Node
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		visible, sz := false, size.Event{}
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					visible = true
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx, sz)
				case lifecycle.CrossOff:
					visible = false
					loadscene = false
					onStop()
				}
			case size.Event:
				sz = e
			case paint.Event:
				onPaint(glctx, sz)
				a.Publish()
				if visible {
					// Keep animating.
					a.Send(paint.Event{})
				}
			case touch.Event:
				if endFlag {
					// 終了していたらタッチで再スタート
					onStart(glctx, sz)
				}
				onTouchEnd(e, sz)
			}
		}
	})
}

func onStart(glctx gl.Context, sz size.Event) {
	endFlag = false
	b = board.New13()
	whichTurn = BLACK
	images = glutil.NewImages(glctx)
	fps = debug.NewFPS(images)
	eng = glsprite.Engine(images)
	loadScene(glctx, sz)
}

func onStop() {
	eng.Release()
	fps.Release()
	images.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 1, 1, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	now := clock.Time(time.Since(startTime) * 60 / time.Second)
	eng.Render(scene, now, sz)
	fps.Draw(sz)
	// androidでonStart時に
	// 画面サイズが取れなかったので
	// onPaint内でもっかい呼ぶ
	if !loadscene {
		loadScene(glctx, sz)
	}
}

func onTouchEnd(e touch.Event, sz size.Event) {
	var (
		offset = 1
		posX   int
		posY   int
		n      *sprite.Node
	)

	//log.Printf("x", touchX/sz.PixelsPerPt)
	//log.Printf("y", touchY/sz.PixelsPerPt)

	posX = int(e.X / sz.PixelsPerPt * 12 / float32(sz.WidthPt))
	posY = int((e.Y/sz.PixelsPerPt - float32((sz.HeightPt-sz.WidthPt)/2)) * 12 / float32(sz.WidthPt))

	// 画面右端のために
	// 画面半分過ぎたらポイントの右側に表示
	if int(e.X) > sz.WidthPx/2+sz.WidthPx/12/2 {
		posX += offset
	}

	// タッチ中に座標が移動しなければ何もしない
	if posX == prevPosX && posY == prevPosY {
		return
	}

	// 盤外
	if posX < 0 || posX > 12 || posY < 0 || posY > 12 {
		return
	}

	switch e.Type.String() {
	case "begin":
		// タッチ開始時に画像を作成して表示
		prevN = newNode()
		if whichTurn == BLACK {
			eng.SetSubTex(prevN, goisiTexs[texBlack])
		} else {
			eng.SetSubTex(prevN, goisiTexs[texWhite])
		}
		eng.SetTransform(prevN, f32.Affine{
			{float32(sz.WidthPx/12) / sz.PixelsPerPt, 0, float32(float32(sz.WidthPx/12*posX)/sz.PixelsPerPt - float32(sz.WidthPx/12)/sz.PixelsPerPt/2)},
			{0, float32(sz.WidthPx/12) / sz.PixelsPerPt, float32(float32(sz.WidthPx/12*posY)/sz.PixelsPerPt-float32(sz.WidthPx/12)/sz.PixelsPerPt/2) + float32((sz.HeightPt-sz.WidthPt)/2)},
		})
	case "move":
		// タッチ中は動かせる
		eng.SetTransform(prevN, f32.Affine{
			{float32(sz.WidthPx/12) / sz.PixelsPerPt, 0, float32(float32(sz.WidthPx/12*posX)/sz.PixelsPerPt - float32(sz.WidthPx/12)/sz.PixelsPerPt/2)},
			{0, float32(sz.WidthPx/12) / sz.PixelsPerPt, float32(float32(sz.WidthPx/12*posY)/sz.PixelsPerPt-float32(sz.WidthPx/12)/sz.PixelsPerPt/2) + float32((sz.HeightPt-sz.WidthPt)/2)},
		})
	case "end":
		// 話したら石を置く
		eng.SetSubTex(prevN, sprite.SubTex{})
		n = newNode()
		if whichTurn == BLACK {
			eng.SetSubTex(n, goisiTexs[texBlack])
		} else {
			eng.SetSubTex(n, goisiTexs[texWhite])
		}
		//posX = int(touchX / sz.PixelsPerPt * 12 / float32(sz.WidthPt))
		//posY = int(touchY / sz.PixelsPerPt * 12 / float32(sz.WidthPt))
		log.Printf("posX", posX)
		log.Printf("posY", posY)

		// 置けるかどうか
		canPut := board.PutPos(b, posX, posY, whichTurn)
		if !canPut {
			return
		}

		eng.SetTransform(n, f32.Affine{
			{float32(sz.WidthPx/12) / sz.PixelsPerPt, 0, float32(float32(sz.WidthPx/12*posX)/sz.PixelsPerPt - float32(sz.WidthPx/12)/sz.PixelsPerPt/2)},
			{0, float32(sz.WidthPx/12) / sz.PixelsPerPt, float32(float32(sz.WidthPx/12*posY)/sz.PixelsPerPt-float32(sz.WidthPx/12)/sz.PixelsPerPt/2) + float32((sz.HeightPt-sz.WidthPt)/2)},
		})

		// 終了判定
		gameEnd := board.GameEnd(b)
		if gameEnd {
			endFlag = true
			return
		}

		// ターン交代
		changeTurn()
	}
}

func changeTurn() {
	if whichTurn == BLACK {
		whichTurn = WHITE
	} else {
		whichTurn = BLACK
	}
}

func newNode() *sprite.Node {
	n := &sprite.Node{}
	eng.Register(n)
	scene.AppendChild(n)
	return n
}

func loadScene(glctx gl.Context, sz size.Event) {
	if sz.WidthPt != 0 && sz.HeightPt != 0 {
		loadscene = true
	}
	texs := loadTextures()
	goisiTexs = loadGoisiTextures(glctx)
	scene = &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	var n *sprite.Node

	log.Println(sz.WidthPt)
	log.Println(sz.HeightPt)

	n = newNode()
	eng.SetSubTex(n, texs[texGoban])
	eng.SetTransform(n, f32.Affine{
		{float32(sz.WidthPt), 0, 0},
		{0, float32(sz.WidthPt), float32((sz.HeightPt - sz.WidthPt) / 2)},
		//	{0, float32(sz.WidthPt), 0},
	})
}

const (
	texGoban = iota
	texWhite
	texBlack
)

func loadTextures() []sprite.SubTex {
	a, err := Asset("assets/goban13_x.png")
	if err != nil {
		log.Fatal(err)
	}
	//	defer a.Close()

	buf := bytes.NewBuffer(a)
	img, _, err := image.Decode(buf)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	return []sprite.SubTex{
		texGoban: sprite.SubTex{t, image.Rect(0, 0, 589, 589)},
	}
}

func loadGoisiTextures(glctx gl.Context) []sprite.SubTex {
	// 透過画像読み込みにOpenGL関数を使う
	glctx.Enable(gl.BLEND)
	glctx.BlendEquation(gl.FUNC_ADD)
	glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	a, err := Asset("assets/goisi13.png")
	if err != nil {
		log.Fatal(err)
	}
	//	defer a.Close()

	buf := bytes.NewBuffer(a)
	img, _, err := image.Decode(buf)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	return []sprite.SubTex{
		texWhite: sprite.SubTex{t, image.Rect(0, 0, 49, 49)},
		texBlack: sprite.SubTex{t, image.Rect(50, 0, 99, 49)},
	}
}
