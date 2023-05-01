package main

import (
	"cyberpuk-80/entity"
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

func (g *Game) InitMenu() {
	menuItems := []entity.Item{
		entity.Item{
			IsActive: true,
			Index:    0,
			Name:     "New game",
			YCord:    380,
			Img:      g.Menu.Resources.NewGame,
			Opts:     &ebiten.DrawImageOptions{},
		},
		entity.Item{
			IsActive: false,
			Index:    1,
			Name:     "Contine",
			YCord:    460,
			Img:      g.Menu.Resources.Continue,
			Opts:     &ebiten.DrawImageOptions{},
		},
		entity.Item{
			IsActive: false,
			Index:    2,
			Name:     "Exit",
			YCord:    540,
			Img:      g.Menu.Resources.Exit,
			Opts:     &ebiten.DrawImageOptions{},
		},
	}
	g.Menu.Items = menuItems
}

func (g *Game) drawMenuItems(screen *ebiten.Image) {
	arrow := g.Menu.Resources.Arrow
	arrowOpts := &ebiten.DrawImageOptions{}
	for i, item := range g.Menu.Items {
		sizeX, _ := item.Img.Size()
		g.Menu.Items[i].Opts = &ebiten.DrawImageOptions{}
		g.Menu.Items[i].Opts.GeoM.Translate(float64((screenWidth-sizeX)/2), float64(item.YCord))
		screen.DrawImage(g.Menu.Items[i].Img, g.Menu.Items[i].Opts)
		if item.IsActive {
			arrowOpts.GeoM.Translate(float64(700), float64(item.YCord))
			screen.DrawImage(arrow, arrowOpts)
		}
	}
	// Значок новой игры
	/*newGameItem := g.Menu.Resources.NewGame
	newGameItemOpts := &ebiten.DrawImageOptions{}
	newGameItemX, _ := newGameItem.Size()
	newGameItemOpts.GeoM.Translate(float64((screenWidth-newGameItemX)/2), 380)
	screen.DrawImage(newGameItem, newGameItemOpts)
	// Значок выхода
	exitGameItem := g.Menu.Resources.Exit
	exitGameItemOpts := &ebiten.DrawImageOptions{}
	exitGameItemX, _ := exitGameItem.Size()
	exitGameItemOpts.GeoM.Translate(float64((screenWidth-exitGameItemX)/2), 460)
	screen.DrawImage(exitGameItem, exitGameItemOpts)
	// Значок продолжить игру (только для меню паузы)
	continueItem := g.Menu.Resources.Continue
	continueItemOpts := &ebiten.DrawImageOptions{}
	continueItemX, _ := continueItem.Size()
	continueItemOpts.GeoM.Translate(float64((screenWidth-continueItemX)/2), 540)
	screen.DrawImage(continueItem, continueItemOpts)*/
}

func (g *Game) switchMenuItem(isPressedUp bool) {
	minMenuItem := uint8(0)
	maxMenuItem := uint8(len(g.Menu.Items) - 1)
	for i, item := range g.Menu.Items {
		// Пробегаемся по всем элементам меню, выясняем какой элемент активен сейчас
		if item.IsActive {
			// Делаем найденный пункт не активным
			g.Menu.Items[i].IsActive = false
			// Проверяем, была ли нажата кнопка вверх или кнопка вниз
			if isPressedUp {
				// Поскольку была нажата клавиша вверх, проверяем не выходит ли текущий пункт меню за границы существования массива
				if minMenuItem <= item.Index-1 && item.Index-1 <= maxMenuItem {
					g.Menu.Items[item.Index-1].IsActive = true
				} else {
					g.Menu.Items[maxMenuItem].IsActive = true
				}
			} else {
				if minMenuItem <= item.Index+1 && item.Index+1 <= maxMenuItem {
					g.Menu.Items[item.Index+1].IsActive = true
				} else {
					g.Menu.Items[minMenuItem].IsActive = true
				}
			}
			// Прерываем цикл после установки нового пункта активным
			break
		}
	}
	time.Sleep(time.Millisecond * 100)
}

func (g *Game) selectItem() {
	for _, item := range g.Menu.Items {
		if item.IsActive {
			switch item.Name {
			case "New game":
				g.startRace()
			case "Contine":
			case "Exit":
				g.gameState = gameStateCloseApp
			}
		}
	}
}
