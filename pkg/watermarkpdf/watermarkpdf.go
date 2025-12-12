// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package watermarkpdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"go.probo.inc/probo/pkg/mail"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	watermarkFontSize       = 90
	watermarkRotationDegree = -55.0
	watermarkOpacity        = 0.1
	watermarkScaleFactor    = 1.0
	fontDPI                 = 80
	charWidthRatio          = 0.6
	lineSpacingRatio        = 1.5
)

var (
	fontSizeRatio = 72.0 / float64(fontDPI)
	fontColor     = color.RGBA{0, 0, 0, 255}
)

func AddConfidentialWithTimestamp(pdfData []byte, email mail.Addr) ([]byte, error) {
	reader := bytes.NewReader(pdfData)

	watermarkLines := []string{
		"Confidential",
		email.String(),
		time.Now().Format("2006-01-02"),
	}

	textImage, err := generateTextImage(watermarkLines)
	if err != nil {
		return nil, fmt.Errorf("cannot generate watermark image: %w", err)
	}

	// Apply rotation before pdfcpu scaling instead of using pdfcpu's rotation API
	// to ensure the scaled watermark covers the full page properly
	imageData, err := rotateImage(textImage, watermarkRotationDegree)
	if err != nil {
		return nil, fmt.Errorf("cannot rotate image: %w", err)
	}

	imageReader := bytes.NewReader(imageData)
	desc := fmt.Sprintf(
		"rotation:0,position:c,opacity:%.1f,scalefactor:%.1f rel",
		watermarkOpacity,
		watermarkScaleFactor,
	)
	watermarkConf, err := api.ImageWatermarkForReader(imageReader, desc, true, false, types.POINTS)
	if err != nil {
		return nil, fmt.Errorf("cannot create watermark from reader: %w", err)
	}

	var buf bytes.Buffer
	err = api.AddWatermarks(reader, &buf, nil, watermarkConf, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot add watermark: %w", err)
	}

	return buf.Bytes(), nil
}

func generateTextImage(lines []string) (*image.RGBA, error) {
	maxLineLength := 0
	for _, line := range lines {
		maxLineLength = max(maxLineLength, len(line))
	}

	charWidth := int(float64(watermarkFontSize) * charWidthRatio)
	charHeight := watermarkFontSize
	lineSpacing := int(float64(charHeight) * lineSpacingRatio)

	textWidth := maxLineLength * charWidth
	textHeight := len(lines)*charHeight + (len(lines)-1)*lineSpacing

	textImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))

	ttf, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("cannot parse font: %w", err)
	}

	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    float64(watermarkFontSize) * fontSizeRatio,
		DPI:     fontDPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create font face: %w", err)
	}

	d := &font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(fontColor),
		Face: face,
	}

	for i, line := range lines {
		y := charHeight + i*(charHeight+lineSpacing)

		lineWidth := d.MeasureString(line)
		centerX := textWidth/2 - int(lineWidth>>6)/2

		d.Dot = fixed.Point26_6{
			X: fixed.I(centerX),
			Y: fixed.I(y),
		}
		d.DrawString(line)
	}

	return textImg, nil
}

func rotateImage(src image.Image, angleDegrees float64) ([]byte, error) {
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	angle := angleDegrees * math.Pi / 180.0
	cosAngle := math.Cos(angle)
	sinAngle := math.Sin(angle)
	cos := math.Abs(cosAngle)
	sin := math.Abs(sinAngle)

	rotatedWidth := int(float64(srcWidth)*cos + float64(srcHeight)*sin)
	rotatedHeight := int(float64(srcWidth)*sin + float64(srcHeight)*cos)

	dst := image.NewRGBA(image.Rect(0, 0, rotatedWidth, rotatedHeight))

	srcCenterX := float64(srcWidth) / 2
	srcCenterY := float64(srcHeight) / 2
	dstCenterX := float64(rotatedWidth) / 2
	dstCenterY := float64(rotatedHeight) / 2

	for y := range rotatedHeight {
		for x := range rotatedWidth {
			fx := float64(x) - dstCenterX
			fy := float64(y) - dstCenterY

			srcX := fx*cosAngle + fy*sinAngle + srcCenterX
			srcY := -fx*sinAngle + fy*cosAngle + srcCenterY

			if srcX >= 0 && srcY >= 0 && int(srcX) < srcWidth && int(srcY) < srcHeight {
				srcColor := src.At(int(srcX), int(srcY))
				dst.Set(x, y, srcColor)
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, fmt.Errorf("cannot encode rotated image: %w", err)
	}

	return buf.Bytes(), nil
}
