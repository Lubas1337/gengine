package physics

import (
	"github.com/go-gl/mathgl/mgl32"
)

const (
	// Константы для физического движка
	DefaultJumpSpeed               = 3.0
	DefaultGravity                 = 12.0
	DefaultPenetrationEpsilonSmall = 0.1
	DefaultPenetrationEpsilonBig   = 1.0
	DefaultAirMovementSuppression  = 0.7
	DefaultFlyingSpeedMultipier    = 2.0
	DefaultPositionHistoryLength   = 20
	DefaultTerminalVelocity        = -10.0
)

// RigidBody содержит физическое состояние сущности
type RigidBody struct {
	OnPositionUpdated func(*RigidBody)
	Collider          *Box
	PositionHistory   []mgl32.Vec3
	TripDistance      float32
	Position          mgl32.Vec3
	Velocity          mgl32.Vec3
	Force             mgl32.Vec3
	Mass              float32
	Width, Height     float32
	Flying            bool
	Grounded          bool

	// Настраиваемые параметры физики
	JumpSpeed               float32
	Gravity                 float32
	PenetrationEpsilonSmall float32
	PenetrationEpsilonBig   float32
	AirMovementSuppression  float32
	FlyingSpeedMultipier    float32
	PositionHistoryLength   int
}

// NewRigidBody создает новое физическое тело с заданной позицией, массой и размерами
func NewRigidBody(position mgl32.Vec3, mass, width, height float32) *RigidBody {
	return &RigidBody{
		Position:        position,
		Mass:            mass,
		Width:           width,
		Height:          height,
		Flying:          false,
		Grounded:        false,
		Force:           mgl32.Vec3{},
		Velocity:        mgl32.Vec3{},
		PositionHistory: make([]mgl32.Vec3, 0),

		// Устанавливаем настраиваемые параметры по умолчанию
		JumpSpeed:               DefaultJumpSpeed,
		Gravity:                 DefaultGravity,
		PenetrationEpsilonSmall: DefaultPenetrationEpsilonSmall,
		PenetrationEpsilonBig:   DefaultPenetrationEpsilonBig,
		AirMovementSuppression:  DefaultAirMovementSuppression,
		FlyingSpeedMultipier:    DefaultFlyingSpeedMultipier,
		PositionHistoryLength:   DefaultPositionHistoryLength,
	}
}

// UpdateCollider обновляет коллайдер на основе текущей позиции
func (r *RigidBody) UpdateCollider() {
	r.Collider = &Box{
		Min: r.Position.Sub(mgl32.Vec3{r.Width / 2, r.Height, r.Width / 2}),
		Max: r.Position.Add(mgl32.Vec3{r.Width / 2, 0, r.Width / 2}),
	}
}

// Move перемещает физическое тело с использованием прямой скорости.
// Принимает опциональный пол и стены для вычисления коллизий.
func (r *RigidBody) Move(movement mgl32.Vec3, ground *Box, ceiling *Box, walls []Box) {
	wasGrounded := r.Grounded

	// Проверяем состояние "на земле"
	if ground != nil {
		// Устанавливаем состояние "на земле" и сбрасываем вертикальную компоненту скорости
		if !wasGrounded {
			// Если только что приземлились, обнуляем вертикальную скорость
			r.Velocity = mgl32.Vec3{r.Velocity.X(), 0.0, r.Velocity.Z()}
		}

		r.Grounded = true

		// Вычисляем коллизию с землей и перемещаем тело вверх
		if r.Collider != nil {
			b, depth := ground.IntersectionY(*r.Collider)
			if b && depth > 0.001 {
				// Увеличиваем запас при подъеме тела над землей - существенно больше, чтобы точно не провалиться
				r.Position = r.Position.Add(mgl32.Vec3{0, depth + 0.2, 0})
				// Обновляем коллайдер после перемещения
				r.UpdateCollider()
			}
		}
	} else {
		// Если нет земли под игроком, устанавливаем состояние "в воздухе"
		r.Grounded = false
	}

	// Если в воздухе, можем уменьшить движение
	if !r.Grounded && !r.Flying {
		movement = movement.Mul(r.AirMovementSuppression)
	}

	// Временный множитель при полете
	if r.Flying {
		movement = movement.Mul(r.FlyingSpeedMultipier)
	}

	// Проверяем столкновения со стенами более точно и пошагово
	if r.Collider != nil && len(walls) > 0 {
		// Вертикальное движение выполняем маленькими шагами для надежности
		// (особенно важно при падении)
		verticalMove := mgl32.Vec3{0, 0, 0}
		if !r.Flying && !r.Grounded {
			verticalMove = mgl32.Vec3{0, r.Velocity.Y(), 0}
		} else if r.Flying {
			verticalMove = mgl32.Vec3{0, movement.Y(), 0}
		}

		// Если есть вертикальное движение, обрабатываем его маленькими шагами
		if verticalMove.Len() > 0 {
			// Количество шагов зависит от скорости падения
			steps := int(mgl32.Abs(verticalMove.Y())*5.0) + 1 // Минимум 1 шаг
			stepMove := verticalMove.Mul(1.0 / float32(steps))

			for i := 0; i < steps; i++ {
				// Проверяем следующую позицию
				tempPos := r.Position.Add(stepMove)
				r.UpdateColliderAtPosition(tempPos)

				// Проверяем коллизии
				hasCollision := false
				for _, wall := range walls {
					if b, _ := r.Collider.Intersection(wall); b {
						hasCollision = true
						break
					}
				}

				// Если нет коллизии, применяем движение
				if !hasCollision {
					r.Position = tempPos
				} else {
					// Иначе останавливаемся и обнуляем скорость
					if r.Velocity.Y() < 0 {
						r.Grounded = true
					}
					r.Velocity = mgl32.Vec3{r.Velocity.X(), 0, r.Velocity.Z()}
					break
				}
			}
		}

		// Горизонтальное движение тоже выполняем пошагово
		horizMove := mgl32.Vec3{movement.X(), 0, movement.Z()}
		if horizMove.Len() > 0.001 {
			// Делим на 4 шага для плавности
			steps := 4
			stepMove := horizMove.Mul(1.0 / float32(steps))

			for i := 0; i < steps; i++ {
				// Проверяем движение по X и Z вместе
				tempPos := r.Position.Add(stepMove)
				r.UpdateColliderAtPosition(tempPos)

				// Проверяем коллизии
				hasCollision := false
				for _, wall := range walls {
					if b, _ := r.Collider.Intersection(wall); b {
						hasCollision = true
						break
					}
				}

				// Если нет коллизии, применяем движение
				if !hasCollision {
					r.Position = tempPos
				} else {
					break // При коллизии дальше не двигаемся
				}
			}
		}

		// Обновляем коллайдер для текущей позиции
		r.UpdateCollider()
	} else {
		// Если нет стен или коллайдера, просто применяем движение
		r.Position = r.Position.Add(movement)

		// Применяем вертикальную составляющую скорости (если не в режиме полета)
		if !r.Flying {
			r.Position = r.Position.Add(mgl32.Vec3{0, r.Velocity.Y(), 0})
		}

		// Обновляем коллайдер
		r.UpdateCollider()
	}
}

// UpdateColliderAtPosition обновляет коллайдер для заданной позиции (для проверок)
func (r *RigidBody) UpdateColliderAtPosition(position mgl32.Vec3) {
	r.Collider = &Box{
		Min: position.Sub(mgl32.Vec3{r.Width / 2, r.Height, r.Width / 2}),
		Max: position.Add(mgl32.Vec3{r.Width / 2, 0, r.Width / 2}),
	}
}

// Jump заставляет тело прыгнуть, устанавливая скорость
func (r *RigidBody) Jump() {
	if r.Grounded {
		r.Velocity = mgl32.Vec3{r.Velocity.X(), r.JumpSpeed, r.Velocity.Z()}
		r.Grounded = false
	}
}

// AppendHistory добавляет текущую позицию в историю
func (r *RigidBody) AppendHistory() {
	if r.PositionHistory == nil {
		r.PositionHistory = make([]mgl32.Vec3, 0)
	}
	r.PositionHistory = append([]mgl32.Vec3{r.Position}, r.PositionHistory...)
	if len(r.PositionHistory) > r.PositionHistoryLength {
		r.PositionHistory = r.PositionHistory[:len(r.PositionHistory)-1]
	}
}

// sign возвращает знак числа
func sign(x float32) float32 {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}
