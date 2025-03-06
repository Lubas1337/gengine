package main

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/user/gengine/game"
	"github.com/user/gengine/window"
)

func main() {
	// Инициализируем окно
	config := window.DefaultConfig()
	config.Title = "Game Engine Example"

	win, err := window.New(config)
	if err != nil {
		log.Fatalf("Ошибка создания окна: %v", err)
	}
	defer win.Terminate()

	// Создаем основной игровой объект
	gameInstance, err := game.NewGame(win)
	if err != nil {
		log.Fatalf("Ошибка инициализации игры: %v", err)
	}
	defer gameInstance.Cleanup()

	// Создаем игрока в начальной позиции
	playerStartPos := mgl32.Vec3{5, 9, 5}
	gameInstance.CreatePlayer(playerStartPos)

	// Настраиваем обработчики ввода
	gameInstance.SetupInputHandlers()

	// Загружаем игровой мир
	gameInstance.LoadWorld()

	// Запускаем игровой цикл
	gameInstance.Start()
}
