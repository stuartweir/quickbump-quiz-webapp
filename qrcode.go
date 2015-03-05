/* QR Code support for QuickBump
 *
 * Depends the QRencode library
 */

package main

import (
    "image"
    "image/color"
    "image/png"
    "io"
    "net/http"
    "net/url"
    "unsafe"
)

// #cgo LDFLAGS: -lqrencode
// #include <qrencode.h>
import "C"

type QRCode struct {
    Version int
    Width   int
    Data    []byte
}

func BuildQrCode(data string) *QRCode {
    // Can you believe how easy this is? Go is awesome!
    q := C.QRcode_encodeString(C.CString(data), 0, C.QR_ECLEVEL_M, C.QR_MODE_8, 1)
    c := &QRCode{
        int(q.version),
        int(q.width),
        C.GoBytes(unsafe.Pointer(q.data), q.width*q.width),
    }
    // Presumably this isn't a tradegy because Go copies the data in the line above ...
    C.QRcode_free(q)
    return c
}

// Encodes the QRCode into a PNG and writes it to the given writer
// The ppm argument specifies the pixels-per-module; the size of the dots
func (q *QRCode) WritePng(w io.Writer, ppm int) error {
    rect := image.Rect(0, 0, ppm*q.Width, ppm*q.Width)
    p := color.Palette{color.White, color.Black}
    img := image.NewPaletted(rect, p)

    for y := 0; y < rect.Max.Y; y++ {
        for x := 0; x < rect.Max.X; x++ {
            z := (x/ppm) + q.Width * (y/ppm)
            pix := q.Data[z]
            value := pix & 1
            img.Set(x, y, p[value])
        }
    }

    return png.Encode(w, img)
}

type QRHandler struct {
}

func (h *QRHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var err error
    data := r.URL.RawQuery
    if data == "" {
        http.Error(w, "No data provided. Include something in the query string", 500)
    } else if data, err = url.QueryUnescape(data); err != nil {
        http.Error(w, err.Error(), 500)
    } else if err := BuildQrCode(string(data)).WritePng(w, 10); err != nil {
        http.Error(w, err.Error(), 500)
    }
}

func init() {
    QRModule = &QRHandler{}
}
