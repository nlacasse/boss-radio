package screen

import (
	"fmt"
	"image"
	"image/draw"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
	"periph.io/x/conn/v3/display"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/sh1106"
	"periph.io/x/devices/v3/ssd1306/image1bit"
)

type Screen struct {
	dev  *sh1106.Dev
	face font.Face

	mu     sync.RWMutex
	buffer [6]string
}

func New() (*Screen, error) {
	bus, err := i2creg.Open("")
	if err != nil {
		return nil, err
	}
	opts := sh1106.Opts{
		Addr:    0x3c,
		Rotated: true,
	}
	dev, err := sh1106.NewI2C(bus, &opts)
	if err != nil {
		return nil, err
	}
	scrn := &Screen{
		dev:  dev,
		face: basicfont.Face7x13,
		//face: inconsolata.Regular8x16,
	}
	scrn.clearLocked()
	return scrn, nil
}

func (s *Screen) Freeze(d time.Duration) {
	s.mu.Lock()
	time.AfterFunc(d, s.mu.Unlock)
}

func (s *Screen) SetText(text [6]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffer = text
}

func (s *Screen) ClearText() {
	s.SetText([6]string{})
}

func (s *Screen) SetTextLine(i int, text string) {
	if i < 0 || i > 5 {
		panic(fmt.Sprintf("invalid line %d", i))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffer[i] = text
}

func (s *Screen) PushText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy(s.buffer[:], s.buffer[1:])
	s.buffer[5] = text
}

func (s *Screen) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearLocked()
}

func (s *Screen) clearLocked() {
	img := image.NewGray(image.Rect(0, 0, 128, 64))
	s.dev.Draw(img.Bounds(), img, image.Point{X: 0, Y: 0})
}

// convert resizes and converts to black and white an image while keeping
// aspect ratio, put it in a centered image of the same size as the display.
func convert(disp display.Drawer, src image.Image) *image1bit.VerticalLSB {
	screenBounds := disp.Bounds()
	size := screenBounds.Size()
	//src = resize(src, size)
	img := image1bit.NewVerticalLSB(screenBounds)
	r := src.Bounds()
	r = r.Add(image.Point{(size.X - r.Max.X) / 2, (size.Y - r.Max.Y) / 2})
	draw.Draw(img, r, src, image.Point{}, draw.Src)
	return img
}

// resize is a simple but fast nearest neighbor implementation.
//
// If you need something better, please use one of the various high quality
// (slower!) Go packages available on github.
func resize(src image.Image, size image.Point) *image.NRGBA {
	srcMax := src.Bounds().Max
	dst := image.NewNRGBA(image.Rectangle{Max: size})
	for y := 0; y < size.Y; y++ {
		sY := (y*srcMax.Y + size.Y/2) / size.Y
		for x := 0; x < size.X; x++ {
			dst.Set(x, y, src.At((x*srcMax.X+size.X/2)/size.X, sY))
		}
	}
	return dst
}

func (s *Screen) DrawImage(src image.Image) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.clearLocked()

	img := convert(s.dev, src)
	s.dev.Draw(s.dev.Bounds(), img, image.Point{})
}

func (s *Screen) Draw() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	img := image.NewGray(image.Rect(0, 0, 128, 64))
	for i, l := range s.buffer {
		dot := fixed.P(0, 10*(1+i))
		dr := font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(image1bit.On),
			Face: s.face,
			Dot:  dot,
		}
		if i == 0 {
			dr.Face = inconsolata.Bold8x16
		}
		dr.DrawString(l)
	}
	s.dev.Draw(img.Bounds(), img, image.Point{X: 0, Y: 0})
}

func (s *Screen) Invert(b bool) {
	s.dev.Invert(b)
}
