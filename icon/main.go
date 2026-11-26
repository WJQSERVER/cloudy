// Command icon makes the app's icons from icon/cat.webp, the artwork of
// Cloudy's cat (1.3.12): a rounded square on black. It cuts the square out
// along its own edge, measured ray by ray, and writes every icon the app,
// its packages and its website use:
//
//   - resources/icon.png and icon-dev.png, the icon of Windows, Linux and
//     older macOS, on the 944 of 1024 the icon had before;
//   - icon/AppIcon.icon/Assets/cat.png, the art full bleed, its corners
//     filled from inside, for the Icon Composer document of macOS 26;
//   - internal/app/brand: icon.png (the about card, the start) and mascot.png, the
//     cat's face in a circle, where a page has nothing to show;
//   - internal/app/tray: tray.png and tray-template.png, drawn here, a cat's
//     head for the macOS menu bar;
//   - website/public: icon.png and favicon.png.
//
// Run it from the repository root: go run ./icon
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/vector"
	"golang.org/x/image/webp"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func save(name string, img image.Image) {
	f, err := os.Create(name)
	must(err)
	must(png.Encode(f, img))
	f.Close()
	fmt.Println(name)
}

var src *image.RGBA
var N int

func isBlack(x, y int) bool {
	c := src.RGBAAt(x, y)
	return int(c.R)+int(c.G)+int(c.B) < 3*50
}

// boundary distance from centre along angle th where art begins (from outside).
func edge(th float64) float64 {
	h := float64(N) / 2
	for r := h * 1.45; r > 0; r -= 0.25 {
		x, y := h+r*math.Cos(th), h+r*math.Sin(th)
		ix, iy := int(x), int(y)
		if ix < 0 || iy < 0 || ix >= N || iy >= N {
			continue
		}
		if !isBlack(ix, iy) {
			return r
		}
	}
	return 0
}

func main() {
	f, err := os.Open("icon/cat.webp")
	must(err)
	im, err := webp.Decode(f)
	must(err)
	N = im.Bounds().Dx()
	src = image.NewRGBA(image.Rect(0, 0, N, N))
	xdraw.Draw(src, src.Bounds(), im, image.Point{}, xdraw.Src)
	h := float64(N) / 2

	// The boundary measured along rays, every 0.05 degree, smoothed.
	const steps = 7200
	rad := make([]float64, steps)
	for i := range rad {
		rad[i] = edge(2 * math.Pi * float64(i) / steps)
	}
	sm := make([]float64, steps)
	for i := range sm {
		var v []float64
		for d := -60; d <= 60; d++ {
			v = append(v, rad[(i+d+steps)%steps])
		}
		// a low percentile: the edge where art surely is
		for a := 1; a < len(v); a++ {
			for b := a; b > 0 && v[b] < v[b-1]; b-- {
				v[b], v[b-1] = v[b-1], v[b]
			}
		}
		sm[i] = v[len(v)/4]
	}
	edgeAt := func(x, y float64) (float64, float64) {
		dx, dy := x-h, y-h
		th := math.Atan2(dy, dx)
		if th < 0 {
			th += 2 * math.Pi
		}
		f := th / (2 * math.Pi) * steps
		i := int(f) % steps
		t := f - math.Floor(f)
		r := sm[i]*(1-t) + sm[(i+1)%steps]*t
		return math.Hypot(dx, dy), r
	}
	const inset = 3.0
	mask := image.NewAlpha(src.Bounds())
	for y := 0; y < N; y++ {
		for x := 0; x < N; x++ {
			d, r := edgeAt(float64(x)+.5, float64(y)+.5)
			v := (r - inset - d) + .5 // 1px ramp
			if v < 0 {
				v = 0
			}
			if v > 1 {
				v = 1
			}
			mask.Pix[y*N+x] = uint8(v*255 + .5)
		}
	}
	// Full-bleed: outside the mask takes the colour of the art inside.
	full := image.NewRGBA(src.Bounds())
	known := make([]bool, N*N)
	for y := 0; y < N; y++ {
		for x := 0; x < N; x++ {
			if d, r := edgeAt(float64(x)+.5, float64(y)+.5); d <= r-inset-1.5 {
				full.SetRGBA(x, y, src.RGBAAt(x, y))
				known[y*N+x] = true
			}
		}
	}
	for pass := 0; pass < 400; pass++ {
		var add []int
		var cols []color.RGBA
		for y := 0; y < N; y++ {
			for x := 0; x < N; x++ {
				if known[y*N+x] {
					continue
				}
				var r, g, b, n int
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						xx, yy := x+dx, y+dy
						if xx < 0 || yy < 0 || xx >= N || yy >= N || !known[yy*N+xx] {
							continue
						}
						c := full.RGBAAt(xx, yy)
						r, g, b, n = r+int(c.R), g+int(c.G), b+int(c.B), n+1
					}
				}
				if n > 0 {
					add = append(add, y*N+x)
					cols = append(cols, color.RGBA{uint8(r / n), uint8(g / n), uint8(b / n), 255})
				}
			}
		}
		if len(add) == 0 {
			break
		}
		for i, p := range add {
			known[p] = true
			full.SetRGBA(p%N, p/N, cols[i])
		}
	}
	// Masked icon: art with alpha.
	cut := image.NewNRGBA(src.Bounds())
	for i := 0; i < N*N; i++ {
		c := full.Pix[i*4 : i*4+4]
		copy(cut.Pix[i*4:], c)
		cut.Pix[i*4+3] = mask.Pix[i]
	}

	scaled := func(img image.Image, size, content int) *image.RGBA {
		out := image.NewRGBA(image.Rect(0, 0, size, size))
		off := (size - content) / 2
		xdraw.CatmullRom.Scale(out, image.Rect(off, off, off+content, off+content), img, img.Bounds(), xdraw.Over, nil)
		return out
	}
	full1024 := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	xdraw.CatmullRom.Scale(full1024, full1024.Bounds(), full, full.Bounds(), xdraw.Src, nil)
	save("icon/AppIcon.icon/Assets/cat.png", full1024)
	icon := scaled(cut, 1024, 944)
	save("resources/icon.png", icon)
	save("resources/icon-dev.png", icon)
	save("website/public/icon.png", icon)
	save("website/public/favicon.png", scaled(cut, 128, 128))
	save("internal/app/brand/icon.png", scaled(cut, 512, 472))
	save("internal/app/tray/tray.png", scaled(cut, 64, 60))

	const ss = 4
	// Mascot: a round crop of the cat's face.
	mcx, mcy, mr := 505.0*float64(N)/1254, 850.0*float64(N)/1254, 395.0*float64(N)/1254
	ms := int(2 * mr)
	m := image.NewNRGBA(image.Rect(0, 0, ms, ms))
	for y := 0; y < ms; y++ {
		for x := 0; x < ms; x++ {
			sx, sy := int(mcx-mr)+x, int(mcy-mr)+y
			c := full.RGBAAt(sx, sy)
			in := 0
			for j := 0; j < ss; j++ {
				for i := 0; i < ss; i++ {
					dx := float64(x) + (float64(i)+.5)/ss - mr
					dy := float64(y) + (float64(j)+.5)/ss - mr
					if dx*dx+dy*dy <= mr*mr {
						in++
					}
				}
			}
			m.SetNRGBA(x, y, color.NRGBA{c.R, c.G, c.B, uint8(in * 255 / (ss * ss))})
		}
	}
	save("internal/app/brand/mascot.png", scaled(m, 256, 256))

	// Tray template for the macOS menu bar: a cat's head, black.
	const T = 36
	z := vector.NewRasterizer(T, T)
	k := float32(T) / 36
	// head: ellipse via bezier approximations
	ell := func(cx, cy, rx, ry float32) {
		const c = 0.5523
		z.MoveTo(cx+rx, cy)
		z.CubeTo(cx+rx, cy+ry*c, cx+rx*c, cy+ry, cx, cy+ry)
		z.CubeTo(cx-rx*c, cy+ry, cx-rx, cy+ry*c, cx-rx, cy)
		z.CubeTo(cx-rx, cy-ry*c, cx-rx*c, cy-ry, cx, cy-ry)
		z.CubeTo(cx+rx*c, cy-ry, cx+rx, cy-ry*c, cx+rx, cy)
		z.ClosePath()
	}
	ell(18*k, 22*k, 13*k, 10.5*k)
	// ears
	z.MoveTo(5.6*k, 19*k)
	z.LineTo(6.8*k, 4.6*k)
	z.LineTo(14.8*k, 12.6*k)
	z.ClosePath()
	z.MoveTo(30.4*k, 19*k)
	z.LineTo(21.2*k, 12.6*k)
	z.LineTo(29.2*k, 4.6*k)
	z.ClosePath()
	al := image.NewAlpha(image.Rect(0, 0, T, T))
	z.Draw(al, al.Bounds(), image.Opaque, image.Point{})
	// eyes cut out
	e := vector.NewRasterizer(T, T)
	ellE := func(cx, cy, rx, ry float32) {
		const c = 0.5523
		e.MoveTo(cx+rx, cy)
		e.CubeTo(cx+rx, cy+ry*c, cx+rx*c, cy+ry, cx, cy+ry)
		e.CubeTo(cx-rx*c, cy+ry, cx-rx, cy+ry*c, cx-rx, cy)
		e.CubeTo(cx-rx, cy-ry*c, cx-rx*c, cy-ry, cx, cy-ry)
		e.CubeTo(cx+rx*c, cy-ry, cx+rx, cy-ry*c, cx+rx, cy)
		e.ClosePath()
	}
	ellE(13.2*k, 21.5*k, 2.1*k, 2.6*k)
	ellE(22.8*k, 21.5*k, 2.1*k, 2.6*k)
	eye := image.NewAlpha(image.Rect(0, 0, T, T))
	e.Draw(eye, eye.Bounds(), image.Opaque, image.Point{})
	tt := image.NewNRGBA(image.Rect(0, 0, T, T))
	for i := range al.Pix {
		v := int(al.Pix[i]) * (255 - int(eye.Pix[i])) / 255
		tt.Pix[i*4+3] = uint8(v)
	}
	save("internal/app/tray/tray-template.png", tt)
}
