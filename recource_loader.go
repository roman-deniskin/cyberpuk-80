package main

import (
	"cyberpuk-80/entity"
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"time"
)

func loadCarImages() (*ebiten.Image, *ebiten.Image, error) {
	car, _, err := ebitenutil.NewImageFromFile("img\\my-car\\dmc-24.png")
	if err != nil {
		return nil, nil, err
	}
	carLights, _, err := ebitenutil.NewImageFromFile("img\\my-car\\dmc-lights-24.png")
	if err != nil {
		return nil, nil, err
	}
	return car, carLights, nil
}

func loadRoadImages() ([]*ebiten.Image, error) {
	roadImages := []*ebiten.Image{}

	files, err := ioutil.ReadDir("img\\road\\jpg")
	if err != nil {
		return nil, err
	}

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

	return roadImages, nil
}

func loadFrontCars() ([]*entity.FrontCar, error) {
	rand.Seed(time.Now().UnixNano())

	colors := [10]string{"blue.png", "dark-orange.png", "dark-red.png", "dark-yellow.png", "green.png", "grey.png", "light-blue.png", "magenta.png", "purple.png", "yellow.png"}
	velocityXValues := []float64{-0.9, -0.1, 0.1, 0.9}

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

		velocityXRandomIndex := rand.Intn(len(velocityXValues))

		car = entity.FrontCar{
			VelocityX:     velocityXValues[velocityXRandomIndex],
			VelocityY:     0.5,
			ScaleX:        0.01,
			ScaleY:        0.01,
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
		roadImages:        roadImages,
		bgmPlayer:         bgmPlayer,
		bgDelayMultiplier: 3,
	}, nil
}
