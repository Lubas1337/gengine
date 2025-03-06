package game

import (
	"fmt"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/user/gengine/physics"
	"github.com/user/gengine/renderer"
	"github.com/user/gengine/window"
	"github.com/user/gengine/world"
)

// Game представляет основной игровой процесс
type Game struct {
	Window        *window.Window
	Renderer      *renderer.Renderer
	World         *world.World
	Player        *Player
	PhysicsEngine *physics.PhysicsEngine

	Running      bool
	LastTime     time.Time
	ShowControls bool // Флаг для отображения управления
}

// Константы для управления игрой
const (
	// Расстояние генерации чанков от центра (было 2)
	ChunkDistance = 1
)

// NewGame создает новую игру
func NewGame(win *window.Window) (*Game, error) {
	// Создаем рендерер
	renderer, err := renderer.NewRenderer(win.GetConfig().Width, win.GetConfig().Height)
	if err != nil {
		return nil, fmt.Errorf("Ошибка создания рендерера: %v", err)
	}

	// Создаем мир
	w := world.NewWorld()

	// Создаем физический движок
	physicsEngine := physics.NewPhysicsEngine()

	// Создаем игру
	g := &Game{
		Window:        win,
		Renderer:      renderer,
		World:         w,
		PhysicsEngine: physicsEngine,
		Running:       false,
		LastTime:      time.Now(),
	}

	// Загружаем мир
	g.LoadWorld()

	// Создаем игрока в центре мира
	g.CreatePlayer(mgl32.Vec3{0, 5, 0})

	// Регистрируем физическое тело игрока в движке
	g.PhysicsEngine.Register(g.Player.Body)

	return g, nil
}

// SetupInputHandlers настраивает обработчики ввода
func (g *Game) SetupInputHandlers() {
	// Настраиваем колбэк для обработки движения мыши
	lastX, lastY := float64(g.Window.GetConfig().Width/2), float64(g.Window.GetConfig().Height/2)
	firstMouse := true

	g.Window.GetGLFWWindow().SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		if firstMouse {
			lastX, lastY = xpos, ypos
			firstMouse = false
			return
		}

		xoffset := xpos - lastX
		yoffset := ypos - lastY
		lastX, lastY = xpos, ypos

		g.Player.Camera.UpdateRotation(xoffset, yoffset)
	})

	// Захватываем курсор
	g.Window.SetCursorMode(glfw.CursorDisabled)
}

// CreatePlayer создает игрока
func (g *Game) CreatePlayer(position mgl32.Vec3) {
	g.Player = NewPlayer(position)
	// Регистрируем тело игрока в физическом движке
	g.PhysicsEngine.Register(g.Player.Body)
}

// LoadWorld загружает игровой мир
func (g *Game) LoadWorld() {
	// Генерируем центральный чанк
	chunk := world.NewChunk(mgl32.Vec3{0, 0, 0})

	// Создаем базовый пол толщиной еще больше для максимальной надежности
	const FloorThickness = 10 // Еще больше увеличиваем толщину пола

	// Базовый ландшафт: земля на уровне 0, с необходимой толщиной
	for i := 0; i < world.ChunkWidth; i++ {
		for j := 0; j < world.ChunkWidth; j++ {
			// Создаем базовую поверхность с толщиной
			for y := 0; y < FloorThickness; y++ {
				chunk.SetBlock(i, y, j, "stone", true)
			}

			// Создаем лестницу, но только в некоторых местах
			if i == 12 && j > 5 && j < 10 {
				for h := FloorThickness; h <= FloorThickness+3; h++ {
					chunk.SetBlock(i, h, j, "brick", true)
				}
			}

			// Создаем небольшую платформу
			if i >= 6 && i <= 9 && j >= 6 && j <= 9 {
				chunk.SetBlock(i, FloorThickness, j, "brick", true)
			}
		}
	}

	// Добавляем чанк в мир
	g.World.AddChunk(chunk)

	// Генерируем только один соседний чанк для максимальной производительности
	// но с очень толстым полом
	chunkPos := mgl32.Vec3{
		float32(world.ChunkWidth),
		0,
		0,
	}

	otherChunk := world.NewChunk(chunkPos)

	// Создаем только базовую землю с толщиной
	for i := 0; i < world.ChunkWidth; i++ {
		for j := 0; j < world.ChunkWidth; j++ {
			for y := 0; y < FloorThickness; y++ {
				otherChunk.SetBlock(i, y, j, "stone", true)
			}
		}
	}

	// Добавляем один куб в центре для ориентации
	centerX := world.ChunkWidth / 2
	centerZ := world.ChunkWidth / 2

	// Куб 3x3x3
	for h := FloorThickness; h < FloorThickness+3; h++ {
		for i := centerX - 1; i <= centerX+1; i++ {
			for j := centerZ - 1; j <= centerZ+1; j++ {
				otherChunk.SetBlock(i, h, j, "brick", true)
			}
		}
	}

	// Добавляем чанк в мир
	g.World.AddChunk(otherChunk)
}

// GetControlKeys возвращает список кнопок управления
func (g *Game) GetControlKeys() []struct{ Key, Desc string } {
	return []struct{ Key, Desc string }{
		{"W", "Движение вперед"},
		{"A", "Движение влево"},
		{"S", "Движение назад"},
		{"D", "Движение вправо"},
		{"Space", "Прыжок / Полет вверх"},
		{"Shift", "Полет вниз (в режиме полета)"},
		{"F", "Переключение режима полета"},
		{"Escape", "Выход из игры"},
		{"H", "Показать/скрыть это меню"},
	}
}

// ProcessInput обрабатывает пользовательский ввод
func (g *Game) ProcessInput() (forward, right, up float32) {
	// Обрабатываем ввод
	if g.Window.IsPressed(glfw.KeyW) {
		forward += 1.0
	}
	if g.Window.IsPressed(glfw.KeyS) {
		forward -= 1.0
	}
	if g.Window.IsPressed(glfw.KeyD) {
		right += 1.0
	}
	if g.Window.IsPressed(glfw.KeyA) {
		right -= 1.0
	}

	// Прыжок
	if g.Window.IsPressed(glfw.KeySpace) && g.Player.OnGround {
		g.Player.Jump()
	}

	// Переключение режима полета - временно отключаем
	// if g.Window.Debounce(glfw.KeyF) {
	// 	// Режим полета временно отключен
	// 	// g.Player.ToggleFlight()
	// }

	// Полет вверх/вниз - временно отключаем
	// if g.Player.IsFlying() {
	// 	if g.Window.IsPressed(glfw.KeySpace) {
	// 		up += 1.0
	// 	}
	// 	if g.Window.IsPressed(glfw.KeyLeftShift) {
	// 		up -= 1.0
	// 	}
	// }

	// Показать/скрыть управление
	if g.Window.Debounce(glfw.KeyH) {
		g.ShowControls = !g.ShowControls
	}

	// Выход из игры
	if g.Window.IsPressed(glfw.KeyEscape) {
		g.Window.GetGLFWWindow().SetShouldClose(true)
	}

	return forward, right, up
}

// UpdatePhysics обновляет физику игры
func (g *Game) UpdatePhysics(delta float64, forward, right, up float32) {
	// Обновляем движение игрока
	if forward != 0 {
		g.Player.MoveForward(float64(forward) * delta)
	}

	if right != 0 {
		g.Player.MoveRight(float64(right) * delta)
	}

	// Обновляем состояние игрока
	g.Player.Update(delta, g.World)
}

// Render отрисовывает текущий кадр
func (g *Game) Render() {
	// Обновляем вид камеры в рендерере
	g.Renderer.SetCamera(
		g.Player.Camera.GetPosition(),
		g.Player.Camera.GetTarget(),
		g.Player.Camera.GetUp(),
	)

	// Начинаем рендеринг
	g.Renderer.Begin()

	// Получаем только видимые чанки для оптимизации рендеринга
	visibleChunks := g.GetVisibleChunks()

	// Отрисовываем только видимые чанки
	for _, chunk := range visibleChunks {
		g.Renderer.DrawChunk(chunk)
	}

	// Отрисовываем коллайдер игрока
	if g.Player.Body.Collider != nil {
		g.Renderer.DrawBox(*g.Player.Body.Collider, mgl32.Vec3{1.0, 0.0, 0.0}) // Красный цвет для игрока
	}

	// Отрисовываем таблицу с управлением
	if g.ShowControls {
		g.Renderer.DrawControls(g.GetControlKeys())
	}

	// Отображаем FPS
	g.Renderer.DrawFPS()

	g.Renderer.End()
}

// GetVisibleChunks возвращает список видимых чанков для оптимизации рендеринга
func (g *Game) GetVisibleChunks() []*world.Chunk {
	// Временно возвращаем все чанки до исправления функции видимости
	return g.World.GetAllChunks()

	/* Отключаем старую реализацию до исправления
	// Позиция и направление камеры
	cameraPos := g.Player.Camera.GetPosition()
	cameraDir := g.Player.Camera.GetFront()

	// Максимальное расстояние видимости чанков (2 чанка)
	visibleDistance := float32(world.ChunkWidth * 2.5)

	visibleChunks := make([]*world.Chunk, 0)

	// Получаем все чанки
	allChunks := g.World.GetAllChunks()

	// Проверяем каждый чанк
	for _, chunk := range allChunks {
		// Получаем центр чанка
		chunkBox := chunk.GetBoundingBox()
		chunkCenter := chunkBox.Min.Add(
			mgl32.Vec3{
				float32(world.ChunkWidth) / 2,
				float32(world.ChunkHeight) / 2,
				float32(world.ChunkWidth) / 2,
			},
		)

		// Вектор от камеры до центра чанка
		toChunk := chunkCenter.Sub(cameraPos)

		// Дистанция до чанка
		distanceToChunk := toChunk.Len()

		// Проверяем расстояние
		if distanceToChunk > visibleDistance {
			// Слишком далеко, пропускаем
			continue
		}

		// Направление к чанку, нормализованное
		dirToChunk := toChunk.Normalize()

		// Косинус угла между направлением камеры и направлением к чанку
		cosAngle := cameraDir.Dot(dirToChunk)

		// Если косинус положительный, чанк перед камерой
		// Угол < 90 градусов -> cos > 0
		if cosAngle > -0.2 { // Немного захватываем боковые чанки
			visibleChunks = append(visibleChunks, chunk)
		}
	}

	return visibleChunks
	*/
}

// Update обновляет состояние игры
func (g *Game) Update(delta float64) {
	// Обрабатываем ввод
	forward, right, up := g.ProcessInput()

	// Обновляем физику
	g.UpdatePhysics(delta, forward, right, up)

	// Обновляем камеру для рендеринга
	g.Renderer.SetCamera(g.Player.Camera.GetPosition(), g.Player.Camera.GetTarget(), g.Player.Camera.GetUp())

	// Отрисовываем мир
	g.Renderer.Begin()

	// Отрисовываем все чанки
	chunks := g.World.GetAllChunks()
	for _, chunk := range chunks {
		g.Renderer.DrawChunk(chunk)
	}

	// Отображаем управление, если включено
	if g.ShowControls {
		g.Renderer.DrawControls(g.GetControlKeys())
	}

	g.Renderer.End()

	// Обновляем окно
	g.Window.Update()
}

// Start запускает игровой цикл
func (g *Game) Start() {
	g.Running = true
	g.LastTime = time.Now()

	// Максимальный шаг времени для физики (в секундах)
	// Делаем крайне маленьким для максимальной надежности
	const maxDeltaTime = 0.016 // Примерно 60 FPS

	// Счетчик FPS для отладки
	frameCount := 0
	lastFPSTime := time.Now()
	displayFPS := 0

	// Основной игровой цикл
	for !g.Window.ShouldClose() && g.Running {
		// Вычисляем дельту времени
		currentTime := time.Now()
		delta := currentTime.Sub(g.LastTime).Seconds()
		g.LastTime = currentTime

		// Считаем FPS
		frameCount++
		if currentTime.Sub(lastFPSTime).Seconds() >= 1.0 {
			displayFPS = frameCount
			frameCount = 0
			lastFPSTime = currentTime
			fmt.Printf("FPS: %d\n", displayFPS)
		}

		// Ограничиваем максимальную дельту времени
		if delta > maxDeltaTime {
			delta = maxDeltaTime
		}

		// Обновляем состояние игры
		// Обрабатываем ввод
		forward, right, up := g.ProcessInput()

		// Обновляем физику
		g.UpdatePhysics(delta, forward, right, up)

		// Отрисовываем сцену
		g.Render()

		// Обновляем окно
		g.Window.Update()

		// Ограничение скорости цикла для стабильности
		runtime.Gosched()

		// Искусственная задержка для стабильности при слишком высоком FPS
		// Если FPS выше 120, добавляем небольшую задержку
		if displayFPS > 120 {
			time.Sleep(2 * time.Millisecond)
		}
	}
}

// Stop останавливает игровой цикл
func (g *Game) Stop() {
	g.Running = false
}

// Cleanup освобождает ресурсы игры
func (g *Game) Cleanup() {
	if g.Renderer != nil {
		g.Renderer.Destroy()
	}
}

// abs возвращает абсолютное значение целого числа
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
