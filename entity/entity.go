package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type Car struct {
	CarRiddingImg  *ebiten.Image
	CarStoppingImg *ebiten.Image
	CarX           float64
	CarBounds      image.Rectangle
}

type OutcomingObjects struct {
	FrontCar []*FrontCar
}

type FrontCar struct {
	WidthOffset   float64
	ScaleX        float64
	ScaleY        float64
	CollisionBox  image.Rectangle
	X, Y          float64
	Speed         float64
	Car           *DMC
	Img           *ebiten.Image
	ImgName       string
	LightsImg     *ebiten.Image
	LightsImgName string
}

type DMC struct {
	Color int
}

type Menu struct {
	isMenuActive bool
	settings     Settings
}

type Settings struct {
}
