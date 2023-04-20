package main

import (
	"cyberpuk-80/entity"
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
	spawnWidth   = 960
	spawnHeight  = 540
)

const (
	gameStatePlaying = iota
	gameStateGameOver
)

type Game struct {
	bgmPlayer            *audio.Player
	roadImages           []*ebiten.Image
	bgIndex              int
	bgDelay              float32
	bgDelayMultiplier    float32
	Car                  entity.Car
	OutcomingObjects     entity.OutcomingObjects
	GamingObjects        entity.OutcomingObjects
	isStopping           bool
	Menu                 entity.Menu
	spawnTimer           float64
	decelerationTimer    float64
	initialDeceleration  float64
	spawnInterval        float64
	decelerationInterval float64
	score                int
	carsOnScreen         map[int]*entity.FrontCar
	gameFont             font.Face
	gameState            int
}

func createCar(x, y, velocityY float64, img *ebiten.Image) entity.FrontCar {
	return entity.FrontCar{
		CollisionBox: img.Bounds(),
		X:            x,
		Y:            y,
		Car:          nil,
		Img:          nil,
		LightsImg:    nil,
	}
}

func (g *Game) Update() error {
	switch g.gameState {
	case gameStatePlaying:
		err := g.UpdatePlaying()
		if err != nil {
			return err
		}
	case gameStateGameOver:
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			g.restart()
		}
	}

	return nil
}

func (g *Game) UpdatePlaying() error {
	g.spawnTimer += ebiten.CurrentTPS()
	g.decelerationTimer += ebiten.CurrentTPS()

	maxXCordCar := float64(screenWidth - g.Car.CarBounds.Max.X)

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("Выход по нажатию на клавишу Escape")
	}
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
		//g.bgmPlayer.Rewind()       // Верните музыку в начало
		g.bgmPlayer.Play() // Запустите музыку
	}

	// Создание машинки
	if g.spawnTimer >= g.spawnInterval {
		originalFrontCar := g.OutcomingObjects.FrontCar[rand.Intn(len(g.OutcomingObjects.FrontCar))]
		g.carsOnScreen[int(time.Now().Unix())] = copyFrontCar(originalFrontCar)

		// Определяем частосту спавна встречек
		if g.spawnInterval > 1960 {
			g.spawnInterval -= 100
		}
		g.spawnTimer = 0
	}

	// Определяем замедление встречных машин
	if g.decelerationTimer >= g.decelerationInterval {
		if g.initialDeceleration > 10000000 {
			g.initialDeceleration -= 10000000
		}

		if g.decelerationInterval > 10 {
			g.decelerationInterval -= 10
		}

		g.decelerationTimer = 0
	}

	for i, car := range g.carsOnScreen {
		scaleCoef := (1 / (spawnHeight / (car.Y - spawnHeight)))

		car.Speed += ebiten.CurrentTPS() / 300
		car.Y = car.Y + scaleCoef + (math.Pow(car.Speed, 5) / g.initialDeceleration)

		car.X = spawnWidth + (car.WidthOffset / (spawnHeight / (car.Y - spawnHeight)))
		car.ScaleX = scaleCoef
		car.ScaleY = scaleCoef
		if car.Y >= screenHeight {
			g.score++
			delete(g.carsOnScreen, i)
		}

		// Проверка столкновения
		carRect := image.Rect(int(g.Car.CarX), screenHeight-g.Car.CarBounds.Max.Y, int(g.Car.CarX)+g.Car.CarBounds.Max.X, screenHeight)

		carWidth, carHeight := car.Img.Size()
		borderRect := image.Rect(
			int(car.X), int(car.Y),
			int(car.X)+int(float64(carWidth)*car.ScaleX),
			int(car.Y)+int(float64(carHeight)*car.ScaleY),
		)

		if carRect.Overlaps(borderRect) {
			g.gameState = gameStateGameOver
		}
	}

	return nil
}

func (g *Game) restart() {
	g.Car.CarX = float64(screenWidth) / 2
	carsOnScreen := make(map[int]*entity.FrontCar)
	g.initialDeceleration = 500000000
	g.spawnInterval = 4900
	g.decelerationInterval = 4900
	g.score = 0
	g.carsOnScreen = carsOnScreen
	g.gameState = gameStatePlaying
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	msg := fmt.Sprintf("GAME OVER\nYOUR SCORE:  %d\nPRESS ENTER", g.score)
	x := (screenWidth - text.BoundString(g.gameFont, msg).Dx()) / 2
	y := screenHeight / 2
	white := color.RGBA{255, 255, 255, 255}
	text.Draw(screen, msg, g.gameFont, x, y, white)
}

func (g *Game) drawScore(screen *ebiten.Image) {
	yellow := color.RGBA{255, 255, 0, 255}                                             // желтый цвет
	text.Draw(screen, strconv.Itoa(g.score), g.gameFont, screenWidth-200, 100, yellow) // отступы 10 пикселей сверху и слева
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.gameState {
	case gameStatePlaying:
		bgImg := g.roadImages[g.bgIndex]

		bgOpts := &ebiten.DrawImageOptions{}
		screen.DrawImage(bgImg, bgOpts)

		carOpts := &ebiten.DrawImageOptions{}
		car := g.Car.CarRiddingImg
		if g.isStopping {
			car = g.Car.CarStoppingImg
		}

		carOpts.GeoM.Translate(g.Car.CarX, float64(screen.Bounds().Dy()-car.Bounds().Dy()))
		screen.DrawImage(car, carOpts)
		g.drawScore(screen)

		for _, frontCar := range g.carsOnScreen {
			borderOpts := &ebiten.DrawImageOptions{}
			borderOpts.GeoM.Scale(frontCar.ScaleX, frontCar.ScaleY)
			borderOpts.GeoM.Translate(frontCar.X, frontCar.Y)
			screen.DrawImage(frontCar.Img, borderOpts)
		}
	case gameStateGameOver:
		g.drawGameOver(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// Создание файла профиля памяти
	f, err := os.Create("mem.pprof")
	if err != nil {
		log.Fatal("Could not create memory profile: ", err)
	}
	defer f.Close()

	startTime := time.Now()

	game, err := NewGame()
	if err != nil {
		log.Fatalf("failed to create game: %v", err)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("Время выполнения функции NewGame:", duration)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("CyberPunk-80")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("failed to run game: %v", err)
	}

	// Запись профиля памяти
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("Could not write memory profile: ", err)
	}
}

func copyFrontCar(original *entity.FrontCar) *entity.FrontCar {
	return &entity.FrontCar{
		ScaleX:        original.ScaleX,
		ScaleY:        original.ScaleY,
		WidthOffset:   original.WidthOffset,
		CollisionBox:  original.CollisionBox,
		X:             original.X,
		Y:             original.Y,
		Speed:         original.Speed,
		Car:           original.Car,
		Img:           original.Img,
		ImgName:       original.ImgName,
		LightsImg:     original.LightsImg,
		LightsImgName: original.LightsImgName,
	}
}
