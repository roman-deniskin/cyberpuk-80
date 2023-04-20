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
	"strconv"
	"time"
)

const (
	screenWidth          = 1920
	screenHeight         = 1080
	spawnWidth           = 960
	spawnHeight          = 540
	initialDeceleration  = 500000000
	spawnInterval        = 4900
	decelerationInterval = 4900
	gameStatePlaying     = iota
	gameStateGameOver
)

type Game struct {
	bgmPlayer            *audio.Player
	roadImages           []*ebiten.Image
	bgIndex              int
	bgDelay              float32
	bgDelayMultiplier    float32
	MainCar              entity.Car
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
	nextCarId            int
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
			g.startRace()
		}
	}

	return nil
}

func (g *Game) UpdatePlaying() error {
	g.spawnTimer += ebiten.CurrentTPS()
	g.decelerationTimer += ebiten.CurrentTPS()

	err := g.controlLogic()
	if err != nil {
		return err
	}

	g.updateBackground()
	g.launchMusicPlayer()
	g.updateOncomingCars()

	return nil
}

func (g *Game) controlLogic() error {
	maxXCordCar := float64(screenWidth - g.MainCar.CarRiddingImg.Bounds().Max.X)

	if ebiten.IsKeyPressed(ebiten.KeyF10) {
		g.bgmPlayer.Seek(time.Minute * 2)
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("Выход по нажатию на клавишу Escape")
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if g.MainCar.CarX > maxXCordCar*0.2 {
			g.MainCar.CarX -= 8
		} else {
			g.MainCar.CarX += 9
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if g.MainCar.CarX < maxXCordCar*0.8 {
			g.MainCar.CarX += 8
		} else {
			g.MainCar.CarX -= 9
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

	return nil
}

func (g *Game) updateBackground() {
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
}

func (g *Game) launchMusicPlayer() {
	if !g.bgmPlayer.IsPlaying() {
		g.bgmPlayer.SetVolume(0.5)
		//g.bgmPlayer.Rewind()
		g.bgmPlayer.Play()
	}
}

func (g *Game) updateOncomingCars() {
	// Создание машинки
	if g.spawnTimer >= g.spawnInterval {
		g.carsOnScreen[g.nextCarId] = createFrontCar(g.OutcomingObjects.FrontCarImages)
		g.nextCarId++
		g.spawnTimer = 0
	}

	// Определяем частосту спавна встречек
	if g.spawnInterval > 1960 {
		g.spawnInterval -= 100
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
		carRect := image.Rect(int(g.MainCar.CarX), screenHeight-g.MainCar.CarRiddingImg.Bounds().Max.Y, int(g.MainCar.CarX)+g.MainCar.CarRiddingImg.Bounds().Max.X, screenHeight)

		carWidth, carHeight := car.Images.Img.Size()
		borderRect := image.Rect(
			int(car.X), int(car.Y),
			int(car.X)+int(float64(carWidth)*car.ScaleX),
			int(car.Y)+int(float64(carHeight)*car.ScaleY),
		)

		if carRect.Overlaps(borderRect) {
			g.gameState = gameStateGameOver
		}
	}
}

func (g *Game) startRace() {
	carsOnScreen := make(map[int]*entity.FrontCar)

	g.bgDelayMultiplier = 3
	g.MainCar.CarX = float64(screenWidth) / 2
	g.initialDeceleration = initialDeceleration
	g.spawnInterval = spawnInterval
	g.decelerationInterval = decelerationInterval
	g.score = 0
	g.carsOnScreen = carsOnScreen
	g.gameState = gameStatePlaying

	if !g.bgmPlayer.IsPlaying() {
		g.bgmPlayer.SetVolume(0.5)
		g.bgmPlayer.Rewind()
		g.bgmPlayer.Play()
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	msg := fmt.Sprintf("GAME OVER\nYOUR SCORE:  %d \nPRESS ENTER", g.score)
	x := (screenWidth - text.BoundString(g.gameFont, msg).Dx()) / 2
	y := screenHeight / 2
	white := color.RGBA{255, 255, 255, 255}
	text.Draw(screen, msg, g.gameFont, x, y, white)
}

func (g *Game) drawScore(screen *ebiten.Image) {
	yellow := color.RGBA{255, 255, 0, 255}
	text.Draw(screen, strconv.Itoa(g.score), g.gameFont, screenWidth-200, 100, yellow)
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.gameState {
	case gameStatePlaying:
		bgImg := g.roadImages[g.bgIndex]

		bgOpts := &ebiten.DrawImageOptions{}
		screen.DrawImage(bgImg, bgOpts)

		carOpts := &ebiten.DrawImageOptions{}
		car := g.MainCar.CarRiddingImg
		if g.isStopping {
			car = g.MainCar.CarStoppingImg
		}

		carOpts.GeoM.Translate(g.MainCar.CarX, float64(screen.Bounds().Dy()-car.Bounds().Dy()))
		screen.DrawImage(car, carOpts)
		g.drawScore(screen)

		borderOpts := &ebiten.DrawImageOptions{}
		for _, frontCar := range g.carsOnScreen {
			borderOpts.GeoM.Reset()
			borderOpts.GeoM.Scale(frontCar.ScaleX, frontCar.ScaleY)
			borderOpts.GeoM.Translate(frontCar.X, frontCar.Y)
			screen.DrawImage(frontCar.Images.Img, borderOpts)
		}
	case gameStateGameOver:
		g.drawGameOver(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game, err := ResourceInit()
	if err != nil {
		log.Fatalf("failed to init resource: %v", err)
	}

	game.startRace()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("CyberPunk-80")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("failed to run game: %v", err)
	}
}

func createFrontCar(frontCarImages []*entity.FrontCarImages) *entity.FrontCar {
	rand.Seed(time.Now().UnixNano())

	carImages := frontCarImages[rand.Intn(len(frontCarImages))]
	widthOffsets := []float64{-1000, -450, 100, 600}
	widthOffsetRandomIndex := rand.Intn(len(widthOffsets))

	return &entity.FrontCar{
		WidthOffset: widthOffsets[widthOffsetRandomIndex],
		X:           screenWidth / 2,
		Y:           screenHeight / 2,
		Images:      carImages,
	}
}
