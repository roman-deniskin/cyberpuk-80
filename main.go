package main

import (
	"cyberpuk-80/entity"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
)

type Game struct {
	bgmPlayer         *audio.Player
	roadImages        []*ebiten.Image
	bgIndex           int
	bgDelay           float32
	bgDelayMultiplier float32
	Car               entity.Car
	OutcomingObjects  entity.OutcomingObjects
	isStopping        bool
	Menu              entity.Menu
}

func createCar(x, y, velocityY float64, img *ebiten.Image) entity.FrontCar {
	return entity.FrontCar{
		VelocityY:    velocityY,
		CollisionBox: img.Bounds(),
		X:            x,
		Y:            y,
		Car:          nil,
		Img:          nil,
		LightsImg:    nil,
	}
}

func (g *Game) Update() error {
	maxXCordCar := float64(screenWidth - g.Car.CarBounds.Max.X)

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if g.Car.CarX > maxXCordCar*0.2 {
			g.Car.CarX -= 8
		} else {
			g.Car.CarX += 9
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if g.Car.CarX < maxXCordCar*0.8 {
			g.Car.CarX += 8
		} else {
			g.Car.CarX -= 9
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if g.bgDelayMultiplier > 1 {
			g.bgDelayMultiplier -= 0.01
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		if g.bgDelayMultiplier < 4 {
			g.bgDelayMultiplier += 0.01
		}
		g.isStopping = true
	} else {
		g.isStopping = false
	}

	g.bgDelay += 1
	if g.bgDelay >= g.bgDelayMultiplier {
		skipFrames := 0
		if g.bgDelayMultiplier < 1 {
			// Ограничиваем пропуск кадров до максимум 1
			skipFrames = int(100.0/g.bgDelayMultiplier) - 100
			if skipFrames > 1 {
				skipFrames = 1
			}
		}
		g.bgIndex = (g.bgIndex + 1 + skipFrames) % len(g.roadImages)
		if g.bgIndex < 0 {
			g.bgIndex += len(g.roadImages)
		}
		g.bgDelay = 0
	}

	if !g.bgmPlayer.IsPlaying() {
		g.bgmPlayer.SetVolume(0.5) // Установите громкость музыки (0.0 - 1.0)
		g.bgmPlayer.Rewind()       // Верните музыку в начало
		g.bgmPlayer.Play()         // Запустите музыку
	}

	for i, car := range g.OutcomingObjects.FrontCar {
		car.Y += car.VelocityY

		// Проверка столкновения
		carRect := image.Rect(int(g.Car.CarX), screenHeight-g.Car.CarBounds.Max.Y, int(g.Car.CarX)+g.Car.CarBounds.Max.X, screenHeight)
		borderRect := image.Rect(
			int(car.X), int(car.Y),
			int(car.X)+int(float64(car.CollisionBox.Max.X)*0.5),
			int(car.Y)+int(float64(car.CollisionBox.Max.Y)*0.5),
		)

		if carRect.Overlaps(borderRect) {
			fmt.Println("Collision detected!")
			// здесь можно обработать столкновение, например, завершить игру
		}

		g.OutcomingObjects.FrontCar[i] = car
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	bgImg := g.roadImages[g.bgIndex]

	bgOpts := &ebiten.DrawImageOptions{}
	screen.DrawImage(bgImg, bgOpts)

	opts := &ebiten.DrawImageOptions{}
	car := g.Car.CarRiddingImg
	if g.isStopping {
		car = g.Car.CarStoppingImg
	}

	for _, frontCar := range g.OutcomingObjects.FrontCar {
		borderOpts := &ebiten.DrawImageOptions{}
		scaleX, scaleY := 0.5, 0.5 // коэффициенты масштабирования по осям X и Y
		borderOpts.GeoM.Scale(scaleX, scaleY)
		borderOpts.GeoM.Translate(frontCar.X, frontCar.Y)
		fmt.Println(frontCar.ImgName)
		screen.DrawImage(frontCar.Img, borderOpts)
	}

	opts.GeoM.Translate(g.Car.CarX, float64(screen.Bounds().Dy()-car.Bounds().Dy()))
	screen.DrawImage(car, opts)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game, err := NewGame()
	if err != nil {
		log.Fatalf("failed to create game: %v", err)
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("CyberPunk-80")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("failed to run game: %v", err)
	}
}
