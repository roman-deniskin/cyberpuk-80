package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
)

type Car struct {
	CarRiddingImg  *ebiten.Image
	CarStoppingImg *ebiten.Image
	CarX           float64
	CarY           float64
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
	Resources    *Resources
}

type Settings struct {
}

type Resources struct {
	Arrow             *ebiten.Image
	Background        *ebiten.Image
	GameOver          *ebiten.Image
	NewGame           *ebiten.Image
	Continue          *ebiten.Image
	Exit              *ebiten.Image
	Score             *ebiten.Image
	YellowtailRegular font.Face
}
