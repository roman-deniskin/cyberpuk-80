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
	"math/rand"
)

const (
	screenWidth   = 1920
	screenHeight  = 1080
	spawnInterval = 9600 // Интервал в кадрах
)

type Game struct {
	bgmPlayer         *audio.Player
	roadImages        []*ebiten.Image
	bgIndex           int
	bgDelay           float32
	bgDelayMultiplier float32
	Car               entity.Car
	OutcomingObjects  entity.OutcomingObjects
	GamingObjects     entity.OutcomingObjects
	isStopping        bool
	Menu              entity.Menu
	spawnTimer        float64
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
	g.spawnTimer += ebiten.CurrentTPS()
	//fmt.Println("interval", g.spawnTimer)

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

	// Создание машинки
	if g.spawnTimer >= spawnInterval {
		fmt.Println("spawn car")
		fmt.Println("gaming objects len: ", len(g.OutcomingObjects.FrontCar))
		originalFrontCar := g.OutcomingObjects.FrontCar[rand.Intn(len(g.OutcomingObjects.FrontCar))]
		newFrontCar := copyFrontCar(originalFrontCar)
		g.GamingObjects.FrontCar = append(g.GamingObjects.FrontCar, &newFrontCar)

		g.spawnTimer = 0
	}

	carVelocityCoef := 0.0
	for i, car := range g.GamingObjects.FrontCar {
		carVelocityCoef = car.Y - screenHeight/2

		var carVelocityX float64
		if car.VelocityX > 0 {
			carVelocityX = car.VelocityX + carVelocityCoef/100
		} else {
			carVelocityX = car.VelocityX - carVelocityCoef/100
		}
		fmt.Println("car velocity", carVelocityX)
		car.Y = car.Y + car.VelocityY + carVelocityCoef/200
		car.X += carVelocityX
		car.ScaleX += carVelocityCoef / 200 / 250
		car.ScaleY += carVelocityCoef / 200 / 250

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

		g.GamingObjects.FrontCar[i] = car
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

	opts.GeoM.Translate(g.Car.CarX, float64(screen.Bounds().Dy()-car.Bounds().Dy()))
	screen.DrawImage(car, opts)

	for _, frontCar := range g.GamingObjects.FrontCar {
		borderOpts := &ebiten.DrawImageOptions{}
		borderOpts.GeoM.Scale(frontCar.ScaleX, frontCar.ScaleY)
		borderOpts.GeoM.Translate(frontCar.X, frontCar.Y)
		//fmt.Println(frontCar.ImgName)
		screen.DrawImage(frontCar.Img, borderOpts)
	}
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

func copyFrontCar(original *entity.FrontCar) entity.FrontCar {
	return entity.FrontCar{
		VelocityX:     original.VelocityX,
		VelocityY:     original.VelocityY,
		ScaleX:        original.ScaleX,
		ScaleY:        original.ScaleY,
		CollisionBox:  original.CollisionBox,
		X:             original.X,
		Y:             original.Y,
		Car:           original.Car,
		Img:           original.Img,
		ImgName:       original.ImgName,
		LightsImg:     original.LightsImg,
		LightsImgName: original.LightsImgName,
	}
}
