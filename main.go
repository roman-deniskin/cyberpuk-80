package main

import (
	"cyberpuk-80/entity"
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
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
	decelerationInterval = 8000
	gameStatePlaying     = iota
	gameStateGameOver
	gameStateCloseApp
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
	YellowtailRegular    font.Face
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
		if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
			g.Menu.KeyUpReleased = true
		}
		if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
			g.Menu.KeyDownReleased = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			g.selectItem()
		}
		if ebiten.IsKeyPressed(ebiten.KeyUp) && g.Menu.KeyUpReleased {
			g.switchMenuItem(true)
			g.Menu.KeyUpReleased = false
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) && g.Menu.KeyDownReleased {
			g.switchMenuItem(false)
			g.Menu.KeyDownReleased = false
		}
	case gameStateCloseApp:
		return fmt.Errorf("It's not an Error, it's just exit")
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
	maxYCordCar := float64(screenHeight - g.MainCar.CarRiddingImg.Bounds().Max.Y)

	if ebiten.IsKeyPressed(ebiten.KeyF10) {
		g.bgmPlayer.Seek(time.Minute * 2)
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("Выход по нажатию на клавишу Escape")
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if g.MainCar.CarX > maxXCordCar*0.13 {
			g.MainCar.CarX -= 20.1
		} else {
			g.MainCar.CarX += 20
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		if g.MainCar.CarX < maxXCordCar*0.87 {
			g.MainCar.CarX += 20.1
		} else {
			g.MainCar.CarX -= 20
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if g.bgDelayMultiplier > 1 {
			g.bgDelayMultiplier -= 0.01
		}
		if g.MainCar.CarY < maxYCordCar*0.20 {
			g.MainCar.CarY += 2
		} else {
			g.MainCar.CarY -= 3
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		if g.bgDelayMultiplier < 4 {
			g.bgDelayMultiplier += 0.01
		}
		if g.MainCar.CarY < maxYCordCar {
			g.MainCar.CarY -= 2
		} else {
			g.MainCar.CarY = maxYCordCar
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

func pixelCollision(img1 *ebiten.Image, x1, y1 int, img2 *ebiten.Image, x2, y2 int) bool {
	width1, height1 := img1.Size()
	width2, height2 := img2.Size()

	for i1 := 0; i1 < width1; i1++ {
		for j1 := 0; j1 < height1; j1++ {
			r1, g1, b1, a1 := img1.At(i1, j1).RGBA()

			if a1 == 0 {
				continue
			}

			i2 := i1 - (x1 - x2)
			j2 := j1 - (y1 - y2)

			if i2 < 0 || i2 >= width2 || j2 < 0 || j2 >= height2 {
				continue
			}

			r2, g2, b2, a2 := img2.At(i2, j2).RGBA()

			if a2 == 0 {
				continue
			}

			if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
				return true
			}
		}
	}

	return false
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

		if g.decelerationInterval > 18 {
			g.decelerationInterval -= 18
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
		//carRect := image.Rect(int(g.MainCar.CarX), screenHeight-g.MainCar.CarRiddingImg.Bounds().Max.Y, int(g.MainCar.CarX)+g.MainCar.CarRiddingImg.Bounds().Max.X, screenHeight)

		//carWidth, carHeight := car.Images.Img.Size()
		/*borderRect := image.Rect(
			int(car.X), int(car.Y),
			int(car.X)+int(float64(carWidth)*car.ScaleX),
			int(car.Y)+int(float64(carHeight)*car.ScaleY),
		)*/

		collision := pixelCollision(
			g.MainCar.CarRiddingImg, int(g.MainCar.CarX), screenHeight-g.MainCar.CarRiddingImg.Bounds().Max.Y,
			car.Images.Img, int(car.X), int(car.Y),
		)

		if collision {
			g.gameState = gameStateGameOver
		}

		/*if carRect.Overlaps(borderRect) {
			g.gameState = gameStateGameOver
		}*/
	}
}

func (g *Game) startRace() {
	carsOnScreen := make(map[int]*entity.FrontCar)

	g.bgDelayMultiplier = 2
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
	bg := g.Menu.Resources.Background
	bgOpts := &ebiten.DrawImageOptions{}
	gameOver := g.Menu.Resources.GameOver
	gameOverOpts := &ebiten.DrawImageOptions{}
	gameOverX, _ := gameOver.Size()
	gameOverOpts.GeoM.Translate(float64((screenWidth-gameOverX)/2), 0)

	screen.DrawImage(bg, bgOpts)
	screen.DrawImage(gameOver, gameOverOpts)
	msg := fmt.Sprintf("Your score: " + strconv.Itoa(g.score))
	yellow := color.RGBA{255, 255, 0, 255}
	text.Draw(screen, msg, g.YellowtailRegular, (screenWidth-text.BoundString(g.YellowtailRegular, msg).Dx())/2, 310, yellow)
	g.drawMenuItems(screen)
}

func (g *Game) drawScore(screen *ebiten.Image) {
	yellow := color.RGBA{255, 255, 0, 255}
	text.Draw(screen, strconv.Itoa(g.score), g.YellowtailRegular, screenWidth-200, 100, yellow)
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
	game.InitMenu()
	game.startRace()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("CyberPunk-80")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeOnlyFullscreenEnabled)
	ebiten.SetTPS(ebiten.SyncWithFPS)
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
