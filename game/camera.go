package game

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	// Константы для управления камерой
	MouseSensitivity = 0.1
	MaxPitch         = 89.0
	MinPitch         = -89.0
)

// Camera представляет собой камеру от первого лица
type Camera struct {
	// Углы Эйлера
	yaw, pitch float64

	// Направление взгляда
	front, right, up mgl32.Vec3

	// Позиция в мире
	position mgl32.Vec3
}

// NewCamera создает новую камеру
func NewCamera(position mgl32.Vec3) *Camera {
	c := &Camera{
		yaw:      -90.0, // Смотрим вдоль отрицательной оси Z
		pitch:    0.0,
		position: position,
		up:       mgl32.Vec3{0, 1, 0},
	}
	c.updateVectors()
	return c
}

// UpdatePosition обновляет позицию камеры
func (c *Camera) UpdatePosition(position mgl32.Vec3) {
	c.position = position
}

// UpdateRotation обновляет углы камеры на основе движения мыши
func (c *Camera) UpdateRotation(xoffset, yoffset float64) {
	xoffset *= MouseSensitivity
	yoffset *= MouseSensitivity

	c.yaw += xoffset
	c.pitch -= yoffset // Инвертируем для корректного управления

	// Ограничиваем углы для предотвращения переворота камеры
	if c.pitch > MaxPitch {
		c.pitch = MaxPitch
	}
	if c.pitch < MinPitch {
		c.pitch = MinPitch
	}

	c.updateVectors()
}

// updateVectors обновляет векторы направления камеры
func (c *Camera) updateVectors() {
	// Вычисляем новый вектор направления взгляда
	radYaw := mgl32.DegToRad(float32(c.yaw))
	radPitch := mgl32.DegToRad(float32(c.pitch))

	c.front = mgl32.Vec3{
		float32(math.Cos(float64(radPitch)) * math.Cos(float64(radYaw))),
		float32(math.Sin(float64(radPitch))),
		float32(math.Cos(float64(radPitch)) * math.Sin(float64(radYaw))),
	}.Normalize()

	// Вычисляем правый и верхний векторы
	c.right = c.front.Cross(mgl32.Vec3{0, 1, 0}).Normalize()
	c.up = c.right.Cross(c.front).Normalize()
}

// GetTarget возвращает точку, на которую смотрит камера
func (c *Camera) GetTarget() mgl32.Vec3 {
	return c.position.Add(c.front)
}

// GetUp возвращает вектор верха камеры
func (c *Camera) GetUp() mgl32.Vec3 {
	return c.up
}

// GetFront возвращает вектор направления взгляда
func (c *Camera) GetFront() mgl32.Vec3 {
	return c.front
}

// GetRight возвращает правый вектор камеры
func (c *Camera) GetRight() mgl32.Vec3 {
	return c.right
}

// GetPosition возвращает текущую позицию камеры
func (c *Camera) GetPosition() mgl32.Vec3 {
	return c.position
}
