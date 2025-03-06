package game

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/user/gengine/physics"
	"github.com/user/gengine/world"
)

// Player представляет игрока в игровом мире
type Player struct {
	Camera     *Camera
	Body       *physics.RigidBody
	Controller *physics.MovementController
	Height     float32
	Width      float32
	OnGround   bool
}

// DefaultPlayerHeight определяет высоту игрока
const DefaultPlayerHeight = 1.8

// DefaultPlayerWidth определяет ширину игрока
const DefaultPlayerWidth = 0.6

// DefaultPlayerMass определяет массу игрока
const DefaultPlayerMass = 70.0

// DefaultPlayerSpeed определяет базовую скорость игрока
const DefaultPlayerSpeed = 1.0

// DefaultJumpForce определяет силу прыжка
const DefaultJumpForce = 4.0

// NewPlayer создает нового игрока
func NewPlayer(position mgl32.Vec3) *Player {
	// Создаем физическое тело с увеличенной устойчивостью
	body := physics.NewRigidBody(position, DefaultPlayerMass, DefaultPlayerWidth, DefaultPlayerHeight)

	// Установка кастомных физических параметров для лучшей игровой физики
	body.JumpSpeed = DefaultJumpForce

	// Создаем контроллер движения с улучшенной скоростью
	controller := physics.NewMovementController(body, DefaultPlayerSpeed, DefaultJumpForce)

	// Создаем камеру на уровне глаз
	cameraPos := position.Add(mgl32.Vec3{0, DefaultPlayerHeight * 0.85, 0})
	camera := NewCamera(cameraPos)

	return &Player{
		Camera:     camera,
		Body:       body,
		Controller: controller,
		Height:     DefaultPlayerHeight,
		Width:      DefaultPlayerWidth,
		OnGround:   false,
	}
}

// Update обновляет состояние игрока
func (p *Player) Update(delta float64, world *world.World) {
	// Обновляем физические параметры и позицию
	p.checkGrounded(world)

	// Обновляем позицию камеры на основе позиции тела
	eyeHeight := p.Height * 0.85 // 85% от высоты для глаз
	p.Camera.UpdatePosition(p.Body.Position.Add(mgl32.Vec3{0, eyeHeight, 0}))
}

// Отдельный метод для проверки касания земли с запасом
func (p *Player) checkGrounded(world *world.World) {
	// Сначала используем стандартное определение "на земле" из физики
	p.OnGround = p.Body.Grounded

	// Дополнительная проверка блоков под игроком
	footPosition := p.Body.Position
	footPosition[1] -= 0.05 // Небольшой запас вниз

	// Проверяем центр и углы под игроком для надежности
	halfWidth := p.Width * 0.45

	// Точки для проверки
	checkPoints := []mgl32.Vec3{
		footPosition, // Центр
		{footPosition.X() - halfWidth, footPosition.Y(), footPosition.Z() - halfWidth},
		{footPosition.X() + halfWidth, footPosition.Y(), footPosition.Z() - halfWidth},
		{footPosition.X() - halfWidth, footPosition.Y(), footPosition.Z() + halfWidth},
		{footPosition.X() + halfWidth, footPosition.Y(), footPosition.Z() + halfWidth},
	}

	// Проверяем все точки
	for _, point := range checkPoints {
		block := world.GetBlock(point)
		if block != nil && block.Active {
			p.OnGround = true
			return
		}
	}
}

// Jump заставляет игрока прыгнуть
func (p *Player) Jump() {
	// Используем и собственную проверку, и проверку из физики
	if p.OnGround || p.Body.Grounded {
		// Дополнительно логируем прыжок для отладки
		fmt.Println("[DEBUG] Игрок прыгнул")
		p.Body.Jump()
		p.OnGround = false
	}
}

// MoveForward перемещает игрока вперед
func (p *Player) MoveForward(amount float64) {
	// Получаем направление "вперед" из камеры, но обнуляем Y
	forward := p.Camera.GetFront()
	forward[1] = 0 // Обнуляем Y для движения по плоскости
	if forward.Len() > 0 {
		forward = forward.Normalize()
	}

	// Применяем движение через контроллер
	p.Controller.Update(float32(amount), 0, 0, forward, p.Camera.GetRight())
}

// MoveRight перемещает игрока вправо
func (p *Player) MoveRight(amount float64) {
	// Получаем направление "вправо" из камеры
	right := p.Camera.GetRight()

	// Применяем движение через контроллер
	p.Controller.Update(0, float32(amount), 0, p.Camera.GetFront(), right)
}

// ProcessMouseMovement обрабатывает движение мыши для камеры
func (p *Player) ProcessMouseMovement(xOffset, yOffset float64, constrainPitch bool) {
	// Просто передаем управление камере
	p.Camera.UpdateRotation(xOffset, yOffset)
}
