package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"net/http"
	"strconv"
)

type mandel struct {
	d        float64 // pixel size in complex plane
	w, h     int     // image size
	n        int     // max number of iterations
	colorful bool
}

type line struct {
	y    int
	data []color.Color
}

// mandel implements image.Image
func (m *mandel) ColorModel() color.Model {
	return color.RGBAModel
}

func (m *mandel) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.w, m.h)
}

func (m *mandel) At(x, y int) color.Color {
	z := complex(float64(x-m.w/2)*m.d, float64(y-m.h/2)*m.d)
	if m.colorful {
		return colorFor(iterate(z, m.n))
	}
	return grayFor(iterate(z, m.n))
}

func (m *mandel) draw1() *image.RGBA {
	img := image.NewRGBA(
		image.Rect(0, 0, m.w, m.h))

	// loop over all lines
	for y := 0; y < m.h; y++ {
		// loop over all pixels over a line
		for x := 0; x < m.w; x++ {
			img.Set(x, y, m.At(x, y))
		}
	}
	return img
}

func (m *mandel) draw2() <-chan *line {
	lines := make(chan *line)
	// loop over all lines
	for y := 0; y < m.h; y++ {
		// launch goroutine for each line
		go func(y int) {
			data := make([]color.Color, m.w)
			for x := range data {
				data[x] = m.At(x, y)
			}
			lines <- &line{y, data}
		}(y)
	}
	return lines
}

func iterate(z complex128, n int) (complex128, int) {
	var v complex128
	for i := 0; i < n; i++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			return v, i
		}
	}
	return v, 0
}

func colorFor(z complex128, n int) color.Color {
	if n == 0 {
		return color.Black
	}
	return color.RGBA{
		R: uint8(100 * cmplx.Abs(z)),
		G: uint8(50 * cmplx.Abs(z)),
		B: uint8(80 * n),
		A: 255,
	}
}

func grayFor(z complex128, n int) color.Color {
	if n == 0 {
		return color.Black
	}
	return color.Gray{uint8(255 - 15*n)}
}

//___________________________________________________________________
func (m *mandel) getFormValues(r *http.Request) {
	m.d = 1 / floatFormValue(r, "zoom", 150)
	m.w = intFormValue(r, "width", 600)
	m.h = intFormValue(r, "height", 600)
	m.n = intFormValue(r, "itertimes", 200)
	m.colorful = boolFormValue(r, "colorful", false)
}

func intFormValue(r *http.Request, key string, val int) int {
	s := r.FormValue(key)
	n, err := strconv.Atoi(s)
	if err != nil {
		n = val
	}
	return n
}

func floatFormValue(r *http.Request, key string, val int) float64 {
	return float64(intFormValue(r, key, val))
}

func boolFormValue(r *http.Request, key string, val bool) bool {
	s := r.FormValue(key)
	b, err := strconv.ParseBool(s)
	if err != nil {
		b = val
	}
	return b
}

func mandelHandler0(w http.ResponseWriter, req *http.Request) {
	var m mandel
	m.getFormValues(req)
	png.Encode(w, &m)
}

func mandelHandler1(w http.ResponseWriter, req *http.Request) {
	var m mandel
	m.getFormValues(req)
	png.Encode(w, m.draw1())
}

func mandelHandler2(w http.ResponseWriter, req *http.Request) {
	var m mandel
	m.getFormValues(req)
	lines := m.draw2()

	img := image.NewRGBA(image.Rect(0, 0, m.w, m.h))
	for i := 0; i < m.h; i++ {
		l := <-lines
		for x, col := range l.data {
			img.Set(x, l.y, col)
		}
	}

	png.Encode(w, img)
}

func main() {
	http.HandleFunc("/mandel0", mandelHandler0)
	http.HandleFunc("/mandel1", mandelHandler1)
	http.HandleFunc("/mandel2", mandelHandler2)
	log.Fatal(http.ListenAndServe(":7777", nil))
}
