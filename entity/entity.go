package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Car struct {
	CarRiddingImg  *ebiten.Image
	CarStoppingImg *ebiten.Image
	CarX           float64
}

type OutcomingObjects struct {
	FrontCarImages []*FrontCarImages
}

type FrontCarImages struct {
	Img       *ebiten.Image
	ImgLights *ebiten.Image
}

type FrontCar struct {
	WidthOffset float64
	ScaleX      float64
	ScaleY      float64
	X, Y        float64
	Speed       float64
	Images      *FrontCarImages
}

type Menu struct {
	isMenuActive bool
	settings     Settings
}

type Settings struct {
}
