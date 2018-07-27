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
	log.Printf("--> %+v", m)
	png.Encode(w, &m)
}

func main() {
	http.HandleFunc("/mandel0", mandelHandler0)
	log.Fatal(http.ListenAndServe(":7777", nil))
}
