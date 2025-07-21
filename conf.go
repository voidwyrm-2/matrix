package main

import (
	"fmt"
	"image"
	"image/color"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/aarzilli/nucular/rect"
	"github.com/aarzilli/nucular/style"
)

type sizeConfig struct {
	Width, Height int
}

func (sc sizeConfig) ToRect() rect.Rect {
	return rect.Rect{0, 0, sc.Width, sc.Height}
}

func (sc sizeConfig) toPoint() image.Point {
	return image.Point{sc.Width, sc.Height}
}

type matrixConfig struct {
	DefaultWindowSize, DefaultPopupSize sizeConfig
	Palette                             matrixPaletteConfig
}

type matrixPaletteConfig struct {
	Text                  string
	Window                string
	Header                string
	HeaderFocused         string
	Border                string
	Button                string
	ButtonHover           string
	ButtonActive          string
	Toggle                string
	ToggleHover           string
	ToggleCursor          string
	Select                string
	SelectActive          string
	Slider                string
	SliderCursor          string
	SliderCursorHover     string
	SliderCursorActive    string
	Property              string
	Edit                  string
	EditCursor            string
	Combo                 string
	Chart                 string
	ChartColor            string
	ChartColorHighlight   string
	Scrollbar             string
	ScrollbarCursor       string
	ScrollbarCursorHover  string
	ScrollbarCursorActive string
	TabHeader             string
}

func (pc matrixPaletteConfig) toTable() (style.ColorTable, error) {
	structValue := reflect.ValueOf(pc)

	rgb := make([]uint8, 0, structValue.NumField()*4)

	for i := range structValue.NumField() {
		hex := structValue.Field(i).String()

		if len(hex) < 6 {
			for range 6 - len(hex) {
				hex += "0"
			}
		} else if len(hex) > 6 {
			return style.ColorTable{}, fmt.Errorf("'%s' is not a valid hexcode (it's too long)", hex)
		}

		pt1, err := strconv.ParseUint(hex[:2], 16, 8)
		if err != nil {
			return style.ColorTable{}, err
		}

		pt2, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return style.ColorTable{}, err
		}

		pt3, err := strconv.ParseUint(hex[4:], 16, 8)
		if err != nil {
			return style.ColorTable{}, err
		}

		rgb = append(rgb, uint8(pt1), uint8(pt2), uint8(pt3), 255)
	}

	arrPtr := unsafe.SliceData(rgb)

	unsPtr := unsafe.Pointer(arrPtr)

	tablePtr := (*style.ColorTable)(unsPtr)

	return *tablePtr, nil
}

var defaultMatrixConfig = matrixConfig{
	DefaultWindowSize: defaultWindowSizeConfig,
	DefaultPopupSize:  defaultPopupSizeConfig,
	Palette:           defaultColorConfig,
}

var (
	defaultWindowSizeConfig = sizeConfig{1100, 1200}
	defaultPopupSizeConfig  = sizeConfig{900, 900}

	defaultColorConfig = matrixPaletteConfig{
		Text:                  "dcdcdc",
		Window:                "142014",
		Header:                "333338",
		HeaderFocused:         "292937",
		Border:                "2e2e2e",
		Button:                "467d64",
		ButtonHover:           "50706e",
		ButtonActive:          "5a647a",
		Toggle:                "323a3d",
		ToggleHover:           "2d3538",
		ToggleCursor:          "30536f",
		Select:                "39433d",
		SelectActive:          "30536f",
		Slider:                "323a3d",
		SliderCursor:          "30536f",
		SliderCursorHover:     "355874",
		SliderCursorActive:    "3a5d79",
		Property:              "323a3d",
		Edit:                  "323a3d",
		EditCursor:            "d2d2d2",
		Combo:                 "323a3d",
		Chart:                 "323a3d",
		ChartColor:            "30536f",
		ChartColorHighlight:   "ff0000",
		Scrollbar:             "323a3d",
		ScrollbarCursor:       "30536f",
		ScrollbarCursorHover:  "355874",
		ScrollbarCursorActive: "3a5d79",
		TabHeader:             "30536f",
	}

	defaultColorTable = style.ColorTable{
		ColorText:                  color.RGBA{220, 220, 220, 255},
		ColorWindow:                color.RGBA{20, 32, 20, 255},
		ColorHeader:                color.RGBA{51, 51, 56, 220},
		ColorHeaderFocused:         color.RGBA{0x29, 0x29, 0x37, 0xdc},
		ColorBorder:                color.RGBA{46, 46, 46, 255},
		ColorButton:                color.RGBA{70, 125, 100, 255},
		ColorButtonHover:           color.RGBA{80, 112, 110, 255},
		ColorButtonActive:          color.RGBA{90, 100, 122, 255},
		ColorToggle:                color.RGBA{50, 58, 61, 255},
		ColorToggleHover:           color.RGBA{45, 53, 56, 255},
		ColorToggleCursor:          color.RGBA{48, 83, 111, 255},
		ColorSelect:                color.RGBA{57, 67, 61, 255},
		ColorSelectActive:          color.RGBA{48, 83, 111, 255},
		ColorSlider:                color.RGBA{50, 58, 61, 255},
		ColorSliderCursor:          color.RGBA{48, 83, 111, 245},
		ColorSliderCursorHover:     color.RGBA{53, 88, 116, 255},
		ColorSliderCursorActive:    color.RGBA{58, 93, 121, 255},
		ColorProperty:              color.RGBA{50, 58, 61, 255},
		ColorEdit:                  color.RGBA{50, 58, 61, 225},
		ColorEditCursor:            color.RGBA{210, 210, 210, 255},
		ColorCombo:                 color.RGBA{50, 58, 61, 255},
		ColorChart:                 color.RGBA{50, 58, 61, 255},
		ColorChartColor:            color.RGBA{48, 83, 111, 255},
		ColorChartColorHighlight:   color.RGBA{255, 0, 0, 255},
		ColorScrollbar:             color.RGBA{50, 58, 61, 255},
		ColorScrollbarCursor:       color.RGBA{48, 83, 111, 255},
		ColorScrollbarCursorHover:  color.RGBA{53, 88, 116, 255},
		ColorScrollbarCursorActive: color.RGBA{58, 93, 121, 255},
		ColorTabHeader:             color.RGBA{48, 83, 111, 255},
	}
)
