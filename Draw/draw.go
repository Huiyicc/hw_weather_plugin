package Draw

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"image"
	"image/color"
	"image/png"
)

var (
	//go:embed static/SourceHanSansCNMedium.ttf
	sourceHanSans []byte

	//go:embed static/qweather-icons.ttf
	qweatherIconsImage []byte

	//go:embed static/icon-table
	icon_table []byte

	//go:embed static/temp.png
	tempPNG []byte
)

var (
	rFont          *truetype.Font
	qweatherFont   *truetype.Font
	qweatherTables map[string]int
)

type Canvas struct {
	ctx *gg.Context
}

func GetRGBA(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}

func initFont() error {
	var err error
	//载入字体数据
	rFont, err = truetype.Parse(sourceHanSans)
	if err != nil {
		return err
	}
	qweatherFont, err = truetype.Parse(qweatherIconsImage)
	if err != nil {
		return err
	}
	qweatherTables = make(map[string]int)
	var tabs []struct {
		Key   string `json:"key"`
		Value int    `json:"value"`
	}
	json.Unmarshal(icon_table, &tabs)
	for _, tab := range tabs {
		qweatherTables[tab.Key] = tab.Value
	}
	return nil
}

// NewCanvas 创建一张图片
//
// width: 宽度
// height: 高度
// background: 背景颜色
func NewCanvas(width, height int, background color.RGBA) (*Canvas, error) {
	if rFont == nil || qweatherFont == nil {
		if err := initFont(); err != nil {
			return nil, err
		}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	dc := gg.NewContextForRGBA(rgba)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.SetRGBA255(int(background.R), int(background.G), int(background.B), int(background.A))
	dc.Fill()
	dc.Clear()
	ret := &Canvas{
		ctx: dc,
	}
	return ret, nil
}

// DrawText 画文字
func (cvs *Canvas) DrawText(str string, size float64, rgba color.Color, left, top int) {
	face := truetype.NewFace(rFont, &truetype.Options{Size: size})
	cvs.ctx.SetFontFace(face)
	cvs.ctx.SetColor(rgba)
	cvs.ctx.DrawStringAnchored(str, float64(left), float64(top), 0, 1)
	cvs.ctx.Fill()
}

// DrawTextVertical 画文字, 从上到下
//
// str: 文字
// size: 字体大小
// rgba: 字体颜色
// left: 左边距
// top: 上边距
func (cvs *Canvas) DrawTextVertical(str string, size float64, rgba color.Color, left, top int) {
	face := truetype.NewFace(rFont, &truetype.Options{Size: size})
	cvs.ctx.SetFontFace(face)
	cvs.ctx.SetColor(rgba)
	l := 0
	for _, val := range str {
		cvs.ctx.DrawStringAnchored(string(val), float64(left), float64(top+(l*int(size))), 0, 1)
		l++
	}
	cvs.ctx.Fill()
}

// DrawWeatherIcon 画天气图标
func (cvs *Canvas) DrawWeatherIcon(index string, size float64, rgba color.Color, left, top int) {
	face := truetype.NewFace(qweatherFont, &truetype.Options{Size: size})
	cvs.ctx.SetFontFace(face)
	cvs.ctx.SetColor(rgba)
	cvs.ctx.DrawStringAnchored(string(rune(qweatherTables[index])), float64(left), float64(top), 0, 1)
	cvs.ctx.Fill()
}

func (cvs *Canvas) DrawBox(left, top, w, h float64, rgba color.Color) {
	cvs.ctx.SetColor(rgba)
	cvs.ctx.DrawRectangle(left, top, w, h)
	cvs.ctx.Fill()
}
func (cvs *Canvas) DrawRoundedBox(left, top, w, h, r float64, rgba color.Color) {
	cvs.ctx.SetColor(rgba)
	cvs.ctx.DrawRoundedRectangle(left, top, w, h, r)
	cvs.ctx.Fill()
}
func (cvs *Canvas) DrawImageData(data []byte, left, top float64) {
	img, _, _ := image.Decode(bytes.NewReader(data))
	cvs.ctx.DrawImageAnchored(img, int(left), int(top), 0, 0)
	cvs.ctx.Fill()
}

// SavePNG 保存图片
func (cvs *Canvas) SavePNG(path string) error {
	return cvs.ctx.SavePNG(path)
}

// SaveBytes 保存图片
func (cvs *Canvas) SaveBytes() ([]byte, error) {
	img := cvs.ctx.Image()
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
