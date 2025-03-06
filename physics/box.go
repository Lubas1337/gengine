package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Box представляет собой AABB (ось-ориентированный ограничивающий бокс) для коллизий
type Box struct {
	Min mgl32.Vec3
	Max mgl32.Vec3
}

// NewBox создает новый бокс по минимальной и максимальной точкам
func NewBox(min, max mgl32.Vec3) Box {
	return Box{
		Min: min,
		Max: max,
	}
}

// Distance вычисляет минимальное расстояние от точки до бокса
func (b Box) Distance(p mgl32.Vec3) float32 {
	// Для каждой оси находим ближайшую точку бокса к заданной точке
	var distSq float32 = 0

	// X-axis
	if p.X() < b.Min.X() {
		distSq += (b.Min.X() - p.X()) * (b.Min.X() - p.X())
	} else if p.X() > b.Max.X() {
		distSq += (p.X() - b.Max.X()) * (p.X() - b.Max.X())
	}

	// Y-axis
	if p.Y() < b.Min.Y() {
		distSq += (b.Min.Y() - p.Y()) * (b.Min.Y() - p.Y())
	} else if p.Y() > b.Max.Y() {
		distSq += (p.Y() - b.Max.Y()) * (p.Y() - b.Max.Y())
	}

	// Z-axis
	if p.Z() < b.Min.Z() {
		distSq += (b.Min.Z() - p.Z()) * (b.Min.Z() - p.Z())
	} else if p.Z() > b.Max.Z() {
		distSq += (p.Z() - b.Max.Z()) * (p.Z() - b.Max.Z())
	}

	return float32(math.Sqrt(float64(distSq)))
}

// CombineY создает новый бокс, объединяющий этот и другой бокс по оси Y
func (b Box) CombineY(other Box) Box {
	return Box{
		Min: mgl32.Vec3{
			b.Min.X(),
			minf(b.Min.Y(), other.Min.Y()),
			b.Min.Z(),
		},
		Max: mgl32.Vec3{
			b.Max.X(),
			maxf(b.Max.Y(), other.Max.Y()),
			b.Max.Z(),
		},
	}
}

// Corners возвращает все 8 углов бокса
func (b Box) Corners() []mgl32.Vec3 {
	return []mgl32.Vec3{
		{b.Min.X(), b.Min.Y(), b.Min.Z()},
		{b.Min.X(), b.Min.Y(), b.Max.Z()},
		{b.Min.X(), b.Max.Y(), b.Min.Z()},
		{b.Min.X(), b.Max.Y(), b.Max.Z()},
		{b.Max.X(), b.Min.Y(), b.Min.Z()},
		{b.Max.X(), b.Min.Y(), b.Max.Z()},
		{b.Max.X(), b.Max.Y(), b.Min.Z()},
		{b.Max.X(), b.Max.Y(), b.Max.Z()},
	}
}

// IntersectionXZ вычисляет пересечение по осям X и Z
func (b Box) IntersectionXZ(other Box) (bool, mgl32.Vec3) {
	// Проверяем пересечение по X и Z
	if b.Max.X() < other.Min.X() || b.Min.X() > other.Max.X() ||
		b.Max.Z() < other.Min.Z() || b.Min.Z() > other.Max.Z() {
		return false, mgl32.Vec3{}
	}

	// Вычисляем глубину проникновения по X
	dx1 := other.Max.X() - b.Min.X()
	dx2 := b.Max.X() - other.Min.X()
	dx := minf(dx1, dx2)

	// Вычисляем глубину проникновения по Z
	dz1 := other.Max.Z() - b.Min.Z()
	dz2 := b.Max.Z() - other.Min.Z()
	dz := minf(dz1, dz2)

	// Выбираем наименьшую глубину проникновения
	var penetration mgl32.Vec3
	if dx < dz {
		penetration = mgl32.Vec3{dx * signF(dx1-dx2), 0, 0}
	} else {
		penetration = mgl32.Vec3{0, 0, dz * signF(dz1-dz2)}
	}

	return true, penetration
}

// IntersectionY вычисляет пересечение по оси Y
func (b Box) IntersectionY(other Box) (bool, float32) {
	// Проверяем пересечение по X и Z
	if b.Max.X() < other.Min.X() || b.Min.X() > other.Max.X() ||
		b.Max.Z() < other.Min.Z() || b.Min.Z() > other.Max.Z() {
		return false, 0
	}

	// Вычисляем глубину проникновения по Y
	if b.Min.Y() > other.Max.Y() || b.Max.Y() < other.Min.Y() {
		return false, 0
	}

	dy1 := other.Max.Y() - b.Min.Y()
	dy2 := b.Max.Y() - other.Min.Y()
	dy := minf(dy1, dy2)

	return true, dy
}

// Intersection проверяет пересечение этого бокса с другим и возвращает глубину проникновения по всем осям
func (b Box) Intersection(other Box) (bool, mgl32.Vec3) {
	// Проверяем пересечение по всем осям
	if b.Max.X() < other.Min.X() || b.Min.X() > other.Max.X() ||
		b.Max.Y() < other.Min.Y() || b.Min.Y() > other.Max.Y() ||
		b.Max.Z() < other.Min.Z() || b.Min.Z() > other.Max.Z() {
		return false, mgl32.Vec3{}
	}

	// Вычисляем глубину проникновения по X
	dx1 := other.Max.X() - b.Min.X()
	dx2 := b.Max.X() - other.Min.X()
	dx := minf(dx1, dx2)

	// Вычисляем глубину проникновения по Y
	dy1 := other.Max.Y() - b.Min.Y()
	dy2 := b.Max.Y() - other.Min.Y()
	dy := minf(dy1, dy2)

	// Вычисляем глубину проникновения по Z
	dz1 := other.Max.Z() - b.Min.Z()
	dz2 := b.Max.Z() - other.Min.Z()
	dz := minf(dz1, dz2)

	// Выбираем наименьшую глубину проникновения
	var penetration mgl32.Vec3
	if dx < dy && dx < dz {
		penetration = mgl32.Vec3{dx * signF(dx1-dx2), 0, 0}
	} else if dy < dx && dy < dz {
		penetration = mgl32.Vec3{0, dy * signF(dy1-dy2), 0}
	} else {
		penetration = mgl32.Vec3{0, 0, dz * signF(dz1-dz2)}
	}

	return true, penetration
}

// minf возвращает минимальное из двух чисел
func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// maxf возвращает максимальное из двух чисел
func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// signF возвращает знак числа как множитель
func signF(x float32) float32 {
	if x < 0 {
		return -1
	}
	return 1
}
