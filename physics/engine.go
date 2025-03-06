package physics

import (
	"github.com/go-gl/mathgl/mgl32"
)

// PhysicsEngine применяет физические вычисления к зарегистрированным RigidBody.
// Метод Tick продвигает симуляцию и вычисляет ускорение, скорость и позицию из приложенных сил.
type PhysicsEngine struct {
	registrations map[*RigidBody]bool
}

// NewPhysicsEngine создает новый физический движок
func NewPhysicsEngine() *PhysicsEngine {
	return &PhysicsEngine{
		registrations: make(map[*RigidBody]bool),
	}
}

// Tick обновляет симуляцию.
// Обновляет все зарегистрированные тела.
func (p *PhysicsEngine) Tick(delta float64) {
	for rb := range p.registrations {
		p.update(rb, delta)
		if rb.OnPositionUpdated != nil {
			rb.OnPositionUpdated(rb)
		}
	}
}

// Register регистрирует RigidBody для обработки на каждом тике.
func (p *PhysicsEngine) Register(body *RigidBody) {
	p.registrations[body] = true
}

// Unregister отменяет регистрацию RigidBody.
func (p *PhysicsEngine) Unregister(body *RigidBody) {
	delete(p.registrations, body)
}

// update обновляет физическое тело с применением физических законов.
func (p *PhysicsEngine) update(body *RigidBody, delta float64) {
	// Обрабатываем гравитацию только если не на земле и не в режиме полета
	if !body.Grounded && !body.Flying {
		// Применяем гравитационную силу
		gravityForce := mgl32.Vec3{0, body.Mass * -body.Gravity, 0}
		body.Force = body.Force.Add(gravityForce)

		// Выводим отладочную информацию
		// fmt.Printf("Гравитация применена: %v, Скорость: %v\n", gravityForce, body.Velocity)
	}

	// Вычисляем ускорение из силы
	acc := body.Force.Mul(1 / body.Mass)

	// Обновляем скорость с учетом ускорения
	body.Velocity = body.Velocity.Add(acc.Mul(float32(delta)))

	// Ограничиваем максимальную скорость падения
	if body.Velocity.Y() < DefaultTerminalVelocity {
		body.Velocity = mgl32.Vec3{body.Velocity.X(), DefaultTerminalVelocity, body.Velocity.Z()}
	}

	// Если на земле, обнуляем вертикальную составляющую скорости
	if body.Grounded && body.Velocity.Y() < 0 {
		body.Velocity = mgl32.Vec3{body.Velocity.X(), 0, body.Velocity.Z()}
	}

	// Вычисляем изменение позиции
	dpos := body.Velocity.Mul(float32(delta))

	// Сохраняем предыдущую позицию в историю
	body.AppendHistory()

	// Обновляем позицию
	body.Position = body.Position.Add(dpos)

	// Обновляем коллайдер
	body.UpdateCollider()

	// Обновляем пройденное расстояние
	body.TripDistance += dpos.Len()

	// Сбрасываем пройденное расстояние, если тело не движется
	if dpos.Len() == 0 && body.TripDistance > 0 {
		body.TripDistance = 0
	}

	// Сбрасываем силу
	body.Force = mgl32.Vec3{}
}
