# GEngine - Игровой движок на Go

GEngine - это легковесный игровой движок, написанный на Go, вдохновленный Minecraft-подобными играми. Он предоставляет базовые компоненты для создания 3D-игр с воксельной графикой.

## Особенности

- **Система окон**: Простая в использовании обертка над GLFW для создания и управления окнами
- **Физический движок**: Реалистичная физика с поддержкой гравитации, коллизий и движения
- **Система чанков**: Эффективное управление миром с помощью чанков для оптимизации производительности
- **Контроллер движения**: Простой в использовании API для управления движением персонажа

## Установка

```bash
go get github.com/user/gengine
```

## Использование

### Создание окна

```go
import "github.com/user/gengine/window"

func main() {
    config := window.DefaultConfig()
    config.Title = "My Game"
    
    win, err := window.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer win.Terminate()
    
    // Основной игровой цикл
    for !win.ShouldClose() {
        // Игровая логика
        
        win.Update()
    }
}
```

### Физика и движение

```go
import (
    "github.com/go-gl/mathgl/mgl32"
    "github.com/user/gengine/physics"
)

// Создаем физический движок
physicsEngine := physics.NewPhysicsEngine()

// Создаем физическое тело
playerPos := mgl32.Vec3{0, 10, 0}
playerBody := physics.NewRigidBody(playerPos, 80.0, 0.6, 1.8)
physicsEngine.Register(playerBody)

// Создаем контроллер движения
movementController := physics.NewMovementController(playerBody, 5.0, 8.0)

// В игровом цикле
delta := 1.0 / 60.0 // или вычислять из реального времени
physicsEngine.Tick(delta)
```

### Работа с миром и чанками

```go
import (
    "github.com/go-gl/mathgl/mgl32"
    "github.com/user/gengine/world"
)

// Создаем мир
gameWorld := world.NewWorld()

// Создаем чанк
chunk := world.NewChunk(mgl32.Vec3{0, 0, 0})

// Добавляем блоки
chunk.SetBlock(0, 0, 0, "stone", true)

// Добавляем чанк в мир
gameWorld.AddChunk(chunk)

// Получаем блок по мировым координатам
block := gameWorld.GetBlock(mgl32.Vec3{1, 1, 1})
```

## Структура проекта

- `window/` - Управление окнами и ввод
- `physics/` - Физический движок и управление движением
- `world/` - Система чанков и управление миром
- `examples/` - Примеры использования библиотеки

## Лицензия

MIT 