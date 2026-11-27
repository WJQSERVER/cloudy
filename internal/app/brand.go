package app

import (
	"bytes"
	_ "embed"
	"image"
	"log"
	"math"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// The license, the source of this fork, and the original upstream, as the
// AGPL asks of a modified version.
const (
	appLicense = appSource + "/blob/main/LICENSE"
	appSource  = "https://github.com/WJQSERVER/cloudy"
	appOrigin  = "https://github.com/egoist/mygo-cloudmusic"
)

//go:embed brand/mascot.png
var mascotPNG []byte

//go:embed brand/icon.png
var brandPNG []byte

func embeddedBitmap(b []byte) func() *ui.Bitmap {
	return sync.OnceValue(func() *ui.Bitmap {
		img, _, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			log.Printf("brand: %v", err)
			return nil
		}
		return ui.NewBitmap(img)
	})
}

// The cat of the app's icon (1.3.12), its face in a circle, and the icon.
var (
	mascot    = embeddedBitmap(mascotPNG)
	brandIcon = embeddedBitmap(brandPNG)
)

// mascotView is who shows where a page has nothing to show.
func mascotView(c *ui.Context, size float32) ui.Element {
	return ui.Image(c, mascot()).Size(size, size).Fit(ui.Cover)
}

// aboutCard says what the app is, and where its license, source and origin
// are. The license itself is the only description here.
func (a *app) aboutCard(k *settingsKit) {
	c := k.c
	ui.Row(c).Padding(16, 20).Gap(14).AlignItems(ui.Center).Children(func() {
		ui.Image(c, brandIcon()).Size(52, 52).Fit(ui.Cover).Shrink(0)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
			ui.Text(c, mygo.App.Name()+" "+mygo.App.Version()).FontSize(14).FontWeight(700)
		})
		if k.button("license", "许可证", iconExternal, false, false).Clicked() {
			openLink(appLicense)
		}
		if k.button("source", "源码", iconExternal, false, false).Clicked() {
			openLink(appSource)
		}
		if k.button("origin", "原项目", iconExternal, false, false).Clicked() {
			openLink(appOrigin)
		}
	})
}

// openLink opens a page in the browser, off the frame.
func openLink(url string) {
	go func() {
		if err := mygo.Shell.OpenExternal(url); err != nil {
			log.Printf("open %s: %v", url, err)
		}
	}()
}

// How the app comes in as it starts (1.3.2, the user's 方案 A): the app's
// icon and a glint crossing it; a record, the cat on its label, growing
// out of it and spinning up as the icon fades; the record flying into the
// player bar's cover as the veil over the app lifts, landing with a ring.
// (Until 1.3.12 the icon was a record player, drawn here, its arm lifting.)
// Times in ms from the start; all of it is painted by one element, by the
// time, so that the frames are painted only, however busy the first page
// is as it loads under it.
const (
	splashEnd   = 1180 * time.Millisecond
	splashTile  = 196 // the icon's side
	splashLiftR = 0.5 // the record's radius risen, of the tile
	// brand/icon.png is brandIconSide pixels, its square brandIconSquare
	// of them (icon/main.go).
	brandIconSide, brandIconSquare = 512, 472
)

// splash is the start of the app (above), over the window. It takes no
// clicks, for nothing to wait on it.
func (a *app) splash(c *ui.Context) {
	if a.splashed {
		return
	}
	if a.motion != motionFull {
		a.splashed = true
		return
	}
	now := c.Now()
	if a.splashAt.IsZero() {
		a.splashAt = now
	}
	gone := now.Sub(a.splashAt)
	if gone >= splashTotal {
		a.splashed = true
		return
	}
	// A build once it is over, to take it away; the frames until then only paint.
	c.After(splashTotal - gone)
	bg, dark := a.pal.bg, a.pal.dark
	ui.Box(c.Key("splash")).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().Draw(func(p *ui.Painter, r ui.Rect) {
		gone := p.Now().Sub(a.splashAt)
		if gone < splashTotal {
			p.AnimationFrame()
		}
		a.drawSplash(p, r, float32(splashTime(gone).Milliseconds()), bg, dark)
	})
}

// The icon, its name under it, stays a while before the record rises off
// it and flies into the bar: 2.5 s in all (1.3.10, the user's choice; the
// 1.18 s of before went by too fast to read it).
const (
	splashHoldAt = 200 * time.Millisecond // the icon in, its name shown
	splashHold   = splashTotal - splashEnd
	splashTotal  = 2500 * time.Millisecond
)

// splashTime is where in the start's course (drawSplash's t) it is, gone
// after it began: held at splashHoldAt for splashHold.
func splashTime(gone time.Duration) time.Duration {
	switch {
	case gone <= splashHoldAt:
		return gone
	case gone <= splashHoldAt+splashHold:
		return splashHoldAt
	}
	return gone - splashHold
}

// drawSplash paints the start at t ms.
func (a *app) drawSplash(p *ui.Painter, r ui.Rect, t float32, bg ui.Color, dark bool) {
	S := float32(splashTile)
	side := a.sideWidth(r.W)
	cx, cy := r.X+side+(r.W-side)/2, r.Y+r.H/2-22
	span := func(from, to float32) float32 { return clamp01((t - from) / (to - from)) }
	out3 := func(x float32) float32 { return 1 - (1-x)*(1-x)*(1-x) }
	inOut3 := func(x float32) float32 {
		if x < .5 {
			return 4 * x * x * x
		}
		y := -2*x + 2
		return 1 - y*y*y/2
	}
	back := func(x float32) float32 { const k = 1.35; y := x - 1; return 1 + (k+1)*y*y*y + k*y*y }

	// The veil over the app, lifting as the record flies, with a warm glow
	// in it about the icon, which goes as the icon does.
	veil := 1 - out3(span(640, 1000))
	if veil > 0 {
		p.Fill(r, bg.Alpha(veil), 0)
		glow := out3(span(0, 260)) * (1 - inOut3(span(380, 760))) * veil
		if glow > 0 {
			g := float32(.16)
			if dark {
				g = .26
			}
			core := ui.Rect{X: cx - 60, Y: cy - 60, W: 120, H: 120}
			p.Shadow(core, 60, 0, 0, 260, 60, ui.RGBA(242, 132, 96, g*glow))
		}
	}

	// The icon, then the record out of it.
	iconIn := span(0, 200) // at rest by the hold
	baseOut := span(380, 660)
	baseAlpha := 1 - out3(baseOut)
	scale := lerp(.6, 1, back(iconIn)) * lerp(1, .9, out3(baseOut))
	if iconIn <= 0 {
		return
	}
	bx, by := cx-S*.1*inOut3(baseOut), cy+S*.04*inOut3(baseOut)
	w := S * scale
	tile := ui.Rect{X: bx - w/2, Y: by - w/2, W: w, H: w}
	lift := inOut3(span(320, 640))
	if baseAlpha > 0 {
		// The icon's square is a squircle: a rounded square a little
		// inside it, its shadow and glint stay under it.
		rad := w * .3
		in := w * .025
		tile := ui.Rect{X: tile.X + in, Y: tile.Y + in, W: w - 2*in, H: w - 2*in}
		p.Shadow(tile, rad, 0, 18, 46, 0, ui.RGBA(120, 70, 40, .30*baseAlpha))
		// The icon's picture has a margin round its square: drawn larger,
		// its square fills the tile.
		if b := brandIcon(); b != nil {
			m := w*(float32(brandIconSide)/brandIconSquare-1)/2 + in
			pic := ui.Rect{X: tile.X - m, Y: tile.Y - m, W: w + 2*m - 2*in, H: w + 2*m - 2*in}
			p.ImageRegion(b, ui.Rect{}, pic, ui.ImageOptions{Tint: white.Alpha(baseAlpha)})
		} else {
			p.Fill(tile, ui.Hex("#e9d2be").Alpha(baseAlpha), rad)
		}
		// A glint crossing it, corner to corner.
		if g := span(180, 460); g > 0 && g < 1 {
			at := lerp(-.3, 1.3, g)
			p.FillGradient(tile, ui.LinearGradient{Angle: 135, Stops: []ui.GradientStop{
				{Color: white.Alpha(0), At: at - .14}, {Color: white.Alpha(.32 * baseAlpha), At: at}, {Color: white.Alpha(0), At: at + .14},
			}}, rad)
		}
	}

	// The record: out of the icon, risen over it, then into the bar.
	ox, oy, oR := bx, by, float32(0)
	ux, uy, uR := cx+S*.06, cy-S*.04, S*splashLiftR
	rx, ry, R := lerp(ox, ux, lift), lerp(oy, uy, lift), lerp(oR, uR, lift)
	shade := lift
	to := a.barSpot.cover
	if to.W <= 0 {
		to = ui.Rect{X: r.X + side + 24, Y: r.Y + r.H - 70, W: 48, H: 48}
	}
	tx, ty, tR := to.X+to.W/2, to.Y+to.H/2, to.W/2
	if fly := span(620, 1000); fly > 0 {
		k := inOut3(fly)
		// An arc: up a little first, then down to the bar.
		c1x, c1y := lerp(rx, tx, .38), min(ry, ty)-80
		ax, ay, bx2, by2 := lerp(rx, c1x, k), lerp(ry, c1y, k), lerp(c1x, tx, k), lerp(c1y, ty, k)
		rx, ry, R = lerp(ax, bx2, k), lerp(ay, by2, k), lerp(R, tR, k)
		shade = lerp(shade, 0, k)
	}
	// Gone into the bar's own cover as it lands.
	if land := span(1000, 1080); land < 1 && R > .5 {
		disc := ui.Rect{X: rx - R, Y: ry - R, W: 2 * R, H: 2 * R}
		p.Shadow(disc, R, 0, R*(.06+.08*shade), R*(.25+.25*shade), 0, ui.RGBA(0, 0, 0, .45*(1-land)))
		if land == 0 {
			drawRecord(p, disc, mascot(), float32(math.Mod(float64(splashTurn(t)), 360)))
		}
	}

	// The name under the icon while it shows.
	if name := out3(span(60, 200)) * (1 - out3(span(380, 560))) * veil; name > 0 {
		ink := ui.RGBA(29, 29, 31, .82*name)
		if dark {
			ink = ui.RGBA(242, 242, 244, .86*name)
		}
		const size = 17
		p.Text(cx-30, cy+S*.5+28+8*(1-out3(span(60, 200))), "Cloudy", size, ink)
	}

	// Landing: a soft ring out of the bar's cover.
	if land := span(1000, 1180); land > 0 && land < 1 {
		g := 18 * out3(land)
		ring := ui.Rect{X: tx - tR - g, Y: ty - tR - g, W: 2 * (tR + g), H: 2 * (tR + g)}
		p.Stroke(ring, ui.RGBA(238, 120, 82, .45*(1-land)), tR+g, 2)
	}
}

// splashTurn is how far, in degrees, the record has turned by t ms: still
// on the icon, spinning up as it rises, easing down to the bar's turn.
func splashTurn(t float32) float32 {
	speed := func(s float32) float32 { // turns a second
		up := 1 - (1-clamp01((s-300)/320))*(1-clamp01((s-300)/320))*(1-clamp01((s-300)/320))
		down := clamp01((s - 760) / 390)
		down = down * down * (3 - 2*down)
		return lerp(up*1.8, .35, down)
	}
	var a float32
	for s := float32(0); s < t; s += 4 {
		a += speed(s) * 4 / 1000 * 360
	}
	return a
}
