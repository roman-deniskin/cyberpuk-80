package main

import (
	"cyberpuk-80/entity"
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

func loadCarImages() (*ebiten.Image, *ebiten.Image, error) {
	car, _, err := ebitenutil.NewImageFromFile("img\\my-car\\dmc-24-mini.png")
	if err != nil {
		return nil, nil, err
	}
	carLights, _, err := ebitenutil.NewImageFromFile("img\\my-car\\dmc-lights-24-mini.png")
	if err != nil {
		return nil, nil, err
	}
	return car, carLights, nil
}

func loadRoadImages() ([]*ebiten.Image, error) {
	// Создание файла профиля памяти
	f, err := os.Create("mem_before_loading_resources.pprof")
	if err != nil {
		log.Fatal("Could not create memory profile: ", err)
	}
	defer f.Close()

	files, err := ioutil.ReadDir("img\\road\\jpg")
	if err != nil {
		return nil, err
	}
	roadImages := make([]*ebiten.Image, 0, len(files))

	for _, file := range files {
		filename := filepath.Join("img\\road\\jpg", file.Name())
		img, _, err := ebitenutil.NewImageFromFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error loading road frame image: %w", err)
		}
		roadImages = append(roadImages, img)
	}

	if len(roadImages) == 0 {
		return nil, errors.New("no road frame images found")
	}

	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("Could not write memory profile: ", err)
	}
	return roadImages, nil
}

func loadFrontCars() ([]*entity.FrontCar, error) {
	rand.Seed(time.Now().UnixNano())

	colors := [10]string{"blue.png", "dark-orange.png", "dark-red.png", "dark-yellow.png", "green.png", "grey.png", "light-blue.png", "magenta.png", "purple.png", "yellow.png"}
	widthOffsets := []float64{-1000, -450, 100, 600}

	var cars []*entity.FrontCar
	for _, color := range colors {
		var car entity.FrontCar
		img, _, err := ebitenutil.NewImageFromFile("img\\front-car\\dmc\\no-lights\\" + color)
		if err != nil {
			return nil, err
		}
		imgLights, _, err := ebitenutil.NewImageFromFile("img\\front-car\\dmc\\lights\\" + color)
		if err != nil {
			return nil, err
		}

		widthOffsetRandomIndex := rand.Intn(len(widthOffsets))

		car = entity.FrontCar{
			WidthOffset:   widthOffsets[widthOffsetRandomIndex],
			CollisionBox:  image.Rectangle{},
			X:             screenWidth / 2,
			Y:             screenHeight / 2,
			Car:           nil,
			Img:           img,
			ImgName:       "img\\front-car\\dmc\\no-lights\\" + color,
			LightsImg:     imgLights,
			LightsImgName: "img\\front-car\\dmc\\lights\\" + color,
		}
		cars = append(cars, &car)
	}
	return cars, nil
}

func loadBackgroundMusic() (*audio.Player, error) {
	// Инициализация аудиосистемы Ebiten с предпочтительным частотой дискретизации.
	audioContext := audio.NewContext(44100)

	// Загрузка аудиофайла.
	rand.Seed(time.Now().UnixNano())
	trackNames := []string{"track1.mp3", "The_Crystal_Method_-_Born_too_Slow.mp3", "Static-X_-_The_Only.mp3", "Snoop_Dogg_The_Doors_-_Riders_On_The_Storm.mp3", "Ying_Yang_Twins_Lil_Jon_The_East_Side_Boyz_-_Get_Low.mp3"}
	trackName := rand.Intn(len(trackNames))
	file, err := ebitenutil.OpenFile("music\\media-player\\" + trackNames[trackName])
	if err != nil {
		return nil, err
	}

	// Декодирование аудиофайла.
	d, err := mp3.Decode(audioContext, file)
	if err != nil {
		file.Close()
		return nil, err
	}

	// Создание плеера для аудиофайла.
	player, err := audio.NewPlayer(audioContext, d)
	// Закрытие файла после декодирования.
	//file.Close()
	if err != nil {
		return nil, err
	}

	return player, nil
}

func loadGameFont() (font.Face, error) {
	fontBytes, err := ioutil.ReadFile("Mario-Kart-DS.ttf")
	if err != nil {
		return nil, err
	}
	tt, err := opentype.Parse(fontBytes) // Замените 'yourFontData' на данные шрифта Mario Kart DS
	if err != nil {
		return nil, err
	}

	const dpi = 72
	return opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48, // размер шрифта
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

// Инициализирует ресурсы при запуске игры
func NewGame() (*Game, error) {
	// Автомобиль игрока
	carRiddingImg, carStoppingImg, err := loadCarImages()
	if err != nil {
		return nil, fmt.Errorf("failed to load car image: %w", err)
	}
	// Трактор встречный
	frontCars, err := loadFrontCars()
	if err != nil {
		return nil, fmt.Errorf("failed to load border image: %w", err)
	}
	roadImages, err := loadRoadImages()
	if err != nil {
		return nil, fmt.Errorf("failed to load road images: %w", err)
	}

	bgmPlayer, err := loadBackgroundMusic()
	if err != nil {
		return nil, fmt.Errorf("failed to load background music: %w", err)
	}
	carBounds := carRiddingImg.Bounds()
	carsOnScreen := make(map[int]*entity.FrontCar)
	gameFont, err := loadGameFont()
	if err != nil {
		return nil, fmt.Errorf("failed to load game font: %w", err)
	}

	return &Game{
		Car: entity.Car{
			CarRiddingImg:  carRiddingImg,
			CarStoppingImg: carStoppingImg,
			CarX:           float64(screenWidth) / 2,
			CarBounds:      carBounds,
		},
		OutcomingObjects: entity.OutcomingObjects{
			FrontCar: frontCars,
		},
		roadImages:           roadImages,
		bgmPlayer:            bgmPlayer,
		bgDelayMultiplier:    3,
		initialDeceleration:  500000000,
		spawnInterval:        4900,
		decelerationInterval: 4900,
		score:                0,
		carsOnScreen:         carsOnScreen,
		gameFont:             gameFont,
	}, nil
}
