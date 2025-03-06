package physics

import (
	"github.com/go-gl/mathgl/mgl32"
)

// MovementController управляет движением персонажа в мире
type MovementController struct {
	Body      *RigidBody
	Speed     float32
	JumpForce float32
	Flying    bool
}

// NewMovementController создает новый контроллер движения
func NewMovementController(body *RigidBody, speed, jumpForce float32) *MovementController {
	return &MovementController{
		Body:      body,
		Speed:     speed,
		JumpForce: jumpForce,
		Flying:    false,
	}
}

// Move перемещает персонажа в направлении, указанном форвардом и боковым движением
func (m *MovementController) Move(forward, right, up float32, viewVector, rightVector mgl32.Vec3) mgl32.Vec3 {
	// Получаем горизонтальные компоненты векторов направления (обнуляем Y)
	flatViewVector := mgl32.Vec3{viewVector.X(), 0, viewVector.Z()}
	if flatViewVector.Len() > 0 {
		flatViewVector = flatViewVector.Normalize()
	}

	flatRightVector := mgl32.Vec3{rightVector.X(), 0, rightVector.Z()}
	if flatRightVector.Len() > 0 {
		flatRightVector = flatRightVector.Normalize()
	}

	// Комбинируем движение в вектор
	movement := flatViewVector.Mul(forward).Add(flatRightVector.Mul(right))

	// Если в воздухе или летим, можем менять компонент по Y
	if m.Flying {
		movement = movement.Add(mgl32.Vec3{0, up, 0})
	}

	// Если длина вектора > 0, нормализуем и умножаем на скорость
	if movement.Len() > 0 {
		// В движении по ровной поверхности нормализуем только для направления
		movement = movement.Normalize().Mul(m.Speed)
	}

	return movement
}

// Jump заставляет персонажа прыгнуть
func (m *MovementController) Jump() {
	if m.Body.Grounded {
		m.Body.Jump()
	}
}

// ToggleFlight переключает режим полета
func (m *MovementController) ToggleFlight() {
	m.Flying = !m.Flying
	m.Body.Flying = m.Flying
}

// Update обновляет состояние контроллера движения
func (m *MovementController) Update(forward, right, up float32, viewVector, rightVector mgl32.Vec3) {
	// Получаем вектор движения
	movement := m.Move(forward, right, up, viewVector, rightVector)

	// Обновляем позицию тела
	m.Body.Position = m.Body.Position.Add(movement)

	// Обновляем историю позиций
	m.Body.AppendHistory()

	// Обновляем коллайдер
	m.Body.UpdateCollider()
}

// GetPosition возвращает текущую позицию
func (m *MovementController) GetPosition() mgl32.Vec3 {
	return m.Body.Position
}

// SetPosition устанавливает новую позицию
func (m *MovementController) SetPosition(pos mgl32.Vec3) {
	m.Body.Position = pos
	m.Body.AppendHistory()
	m.Body.UpdateCollider()
}
