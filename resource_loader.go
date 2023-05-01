package main

import (
	"cyberpuk-80/entity"
	"cyberpuk-80/utils"
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func loadCarImages() (*ebiten.Image, *ebiten.Image, error) {
	startTime := time.Now()

	car, _, err := ebitenutil.NewImageFromFile("img\\my-car\\dmc-24-mini.png")
	if err != nil {
		return nil, nil, err
	}
	carLights, _, err := ebitenutil.NewImageFromFile("img\\my-car\\dmc-lights-24-mini.png")
	if err != nil {
		return nil, nil, err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("Время выполнения функции loadCarImages:", duration)

	return car, carLights, nil
}

func loadRoadImages() ([]*ebiten.Image, error) {
	startTime := time.Now()

	files, err := ioutil.ReadDir("img\\road\\jpg")
	if err != nil {
		return nil, err
	}

	roadImages := make([]*ebiten.Image, len(files))

	type fileWithIndex struct {
		index int
		file  os.FileInfo
	}

	type result struct {
		index int
		img   *ebiten.Image
		err   error
	}

	workerCount := 4
	workChan := make(chan fileWithIndex, len(files))
	resultChan := make(chan result, len(files))

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for work := range workChan {
				filename := filepath.Join("img\\road\\jpg", work.file.Name())
				img, _, err := ebitenutil.NewImageFromFile(filename)
				resultChan <- result{index: work.index, img: img, err: err}
			}
		}()
	}

	go func() {
		for i, file := range files {
			workChan <- fileWithIndex{index: i, file: file}
		}
		close(workChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		if res.err != nil {
			err = res.err
			utils.MessageBox("Error", err.Error(), utils.MB_ICONERROR)
		}
		roadImages[res.index] = res.img
	}

	if err != nil {
		return nil, err
	}

	if len(roadImages) == 0 {
		return nil, errors.New("no road frame images found")
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("Время выполнения функции loadRoadImages:", duration)

	return roadImages, nil
}

func loadFrontCarImages() ([]*entity.FrontCarImages, error) {
	startTime := time.Now()

	colors := [10]string{"blue.png", "dark-orange.png", "dark-red.png", "dark-yellow.png", "green.png", "grey.png", "light-blue.png", "magenta.png", "purple.png", "yellow.png"}

	var cars []*entity.FrontCarImages
	for _, color := range colors {
		var car entity.FrontCarImages
		img, _, err := ebitenutil.NewImageFromFile("img\\front-car\\dmc\\no-lights\\" + color)
		if err != nil {
			return nil, err
		}
		imgLights, _, err := ebitenutil.NewImageFromFile("img\\front-car\\dmc\\lights\\" + color)
		if err != nil {
			return nil, err
		}

		car = entity.FrontCarImages{
			Img:       img,
			ImgLights: imgLights,
		}
		cars = append(cars, &car)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("Время выполнения функции loadFrontCars:", duration)

	return cars, nil
}

func loadMenuResources() (entity.Resources, error) {
	var rec entity.Resources
	var err error
	rec.Background, _, err = ebitenutil.NewImageFromFile("img\\menu\\background.png")
	if err != nil {
		return entity.Resources{}, err
	}
	rec.GameOver, _, err = ebitenutil.NewImageFromFile("img\\menu\\game-over.png")
	if err != nil {
		return entity.Resources{}, err
	}
	rec.Exit, _, err = ebitenutil.NewImageFromFile("img\\menu\\exit-game.png")
	if err != nil {
		return entity.Resources{}, err
	}
	rec.NewGame, _, err = ebitenutil.NewImageFromFile("img\\menu\\new-game.png")
	if err != nil {
		return entity.Resources{}, err
	}
	rec.Continue, _, err = ebitenutil.NewImageFromFile("img\\menu\\continue.png")
	if err != nil {
		return entity.Resources{}, err
	}
	rec.Arrow, _, err = ebitenutil.NewImageFromFile("img\\menu\\arrow.png")
	if err != nil {
		return entity.Resources{}, err
	}
	rec.Score, _, err = ebitenutil.NewImageFromFile("img\\menu\\score.png")
	if err != nil {
		return entity.Resources{}, err
	}
	return rec, nil
}

// Тип используется для музыкального плеера
type ConcatReader struct {
	songs []io.Reader
	index int
}

func (c *ConcatReader) Seek(offset int64, whence int) (int64, error) {
	if whence != 0 {
		return int64(whence) * offset, nil
	}
	// Я хз зачем это нужно) Это походу количество миллисекунд, которые мы скипаем.
	return 1 * offset, nil
}

// Используется для музыкального плеера
func (c *ConcatReader) Read(p []byte) (n int, err error) {
	if c.index >= len(c.songs) {
		return 0, io.EOF
	}

	n, err = c.songs[c.index].Read(p)
	if err == io.EOF {
		c.index++
		return c.Read(p)
	}
	return
}

func loadBackgroundMusic() (*audio.Player, error) {
	startTime := time.Now()

	audioContext := audio.NewContext(22050)

	rand.Seed(time.Now().UnixNano())
	trackFiles, err := ioutil.ReadDir("music\\media-player\\")
	if err != nil {
		return nil, err
	}

	var songs []io.Reader
	for _, trackFile := range trackFiles {
		file, err := ebitenutil.OpenFile("music\\media-player\\" + trackFile.Name())
		if err != nil {
			return nil, err
		}
		song, err := mp3.Decode(audioContext, file)
		if err != nil {
			file.Close()
			return nil, err
		}
		//file.Close()
		songs = append(songs, song)
	}

	audioStream := &ConcatReader{songs: songs}
	player, err := audioContext.NewPlayer(audioStream)
	if err != nil {
		return nil, err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("Время выполнения функции loadBackgroundMusic:", duration)

	return player, nil
}

func loadGameFont() (font.Face, error) {
	startTime := time.Now()

	fontBytes, err := ioutil.ReadFile("Mario-Kart-DS.ttf")
	if err != nil {
		return nil, err
	}
	tt, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	const dpi = 72

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("Время выполнения функции loadGameFont:", duration)

	return opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48, // размер шрифта
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

func loadMenuFont() (font.Face, error) {
	startTime := time.Now()

	fontBytes, err := ioutil.ReadFile("Yellowtail-Regular.ttf")
	if err != nil {
		return nil, err
	}
	tt, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	const dpi = 72

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	fmt.Println("Время выполнения функции loadMenuFont:", duration)

	return opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    72, // размер шрифта
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

func loadIcon() {
	// Загружаем иконку из файла.
	iconFile, err := os.Open("img\\icon.png")
	if err != nil {
		utils.MessageBox("Error", err.Error(), utils.MB_ICONERROR)
	}
	defer iconFile.Close()

	iconImage, _, err := image.Decode(iconFile)
	if err != nil {
		utils.MessageBox("Error", err.Error(), utils.MB_ICONERROR)
	}

	// Устанавливаем иконку окна.
	ebiten.SetWindowIcon([]image.Image{iconImage})
}

func ResourceInit() (*Game, error) {
	var carRiddingImg, carStoppingImg *ebiten.Image
	var frontCarImages []*entity.FrontCarImages
	var roadImages []*ebiten.Image
	var bgmPlayer *audio.Player
	var gameFont font.Face
	var YellowtailRegular font.Face
	var recources entity.Resources

	var loadErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		carRiddingImg, carStoppingImg, loadErr = loadCarImages()
		if loadErr != nil {
			loadErr = fmt.Errorf("failed to load car image: %w", loadErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		frontCarImages, loadErr = loadFrontCarImages()
		if loadErr != nil {
			loadErr = fmt.Errorf("failed to load border image: %w", loadErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		roadImages, loadErr = loadRoadImages()
		if loadErr != nil {
			loadErr = fmt.Errorf("failed to load road images: %w", loadErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bgmPlayer, loadErr = loadBackgroundMusic()
		if loadErr != nil {
			loadErr = fmt.Errorf("failed to load background music: %w", loadErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		gameFont, loadErr = loadGameFont()
		if loadErr != nil {
			loadErr = fmt.Errorf("failed to load game font: %w", loadErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		recources, loadErr = loadMenuResources()
		if loadErr != nil {
			loadErr = fmt.Errorf("failed to load menu recources: %w", loadErr)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		YellowtailRegular, loadErr = loadMenuFont()
		if loadErr != nil {
			loadErr = fmt.Errorf("failed to load menu font: %w", loadErr)
		}
	}()

	wg.Wait()

	if loadErr != nil {
		utils.MessageBox("Error", loadErr.Error(), utils.MB_ICONERROR)
		return nil, loadErr
	}

	return &Game{
		MainCar: entity.Car{
			CarRiddingImg:  carRiddingImg,
			CarStoppingImg: carStoppingImg,
		},
		OutcomingObjects: entity.OutcomingObjects{
			FrontCarImages: frontCarImages,
		},
		roadImages:        roadImages,
		bgmPlayer:         bgmPlayer,
		gameFont:          gameFont,
		YellowtailRegular: YellowtailRegular,
		Menu: entity.Menu{
			Resources:       &recources,
			KeyUpReleased:   true,
			KeyDownReleased: true,
		},
	}, nil
}
