package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/user/gengine/physics"
)

// ChunkSize определяет размер чанка
const (
	ChunkWidth  = 16
	ChunkHeight = 256
)

// BlockData представляет основные данные блока в чанке
type BlockData struct {
	Active    bool
	BlockType string
	Position  mgl32.Vec3
}

// Chunk группирует блоки для рендеринга и операций
type Chunk struct {
	// Блоки в чанке, позиция определяется индексом в массиве
	Blocks [ChunkWidth][ChunkHeight][ChunkWidth]*BlockData

	// Позиция чанка в мире (угол)
	Position mgl32.Vec3
}

// NewChunk создает новый чанк с заданной позицией
func NewChunk(pos mgl32.Vec3) *Chunk {
	c := &Chunk{
		Position: pos,
	}

	// Инициализируем все блоки как неактивные
	for i := 0; i < ChunkWidth; i++ {
		for j := 0; j < ChunkHeight; j++ {
			for k := 0; k < ChunkWidth; k++ {
				c.Blocks[i][j][k] = &BlockData{
					Active:    false,
					BlockType: "",
					Position: mgl32.Vec3{
						pos.X() + float32(i),
						pos.Y() + float32(j),
						pos.Z() + float32(k),
					},
				}
			}
		}
	}

	return c
}

// GetBlock возвращает блок по локальным координатам чанка
func (c *Chunk) GetBlock(x, y, z int) *BlockData {
	if x < 0 || x >= ChunkWidth || y < 0 || y >= ChunkHeight || z < 0 || z >= ChunkWidth {
		return nil
	}
	return c.Blocks[x][y][z]
}

// SetBlock устанавливает блок по локальным координатам чанка
func (c *Chunk) SetBlock(x, y, z int, blockType string, active bool) {
	if x < 0 || x >= ChunkWidth || y < 0 || y >= ChunkHeight || z < 0 || z >= ChunkWidth {
		return
	}
	c.Blocks[x][y][z].BlockType = blockType
	c.Blocks[x][y][z].Active = active
}

// GetBlockFromWorldPos возвращает блок по мировым координатам
func (c *Chunk) GetBlockFromWorldPos(pos mgl32.Vec3) *BlockData {
	// Вычисляем локальные координаты блока внутри чанка
	localX := int(pos.X() - c.Position.X())
	localY := int(pos.Y() - c.Position.Y())
	localZ := int(pos.Z() - c.Position.Z())

	return c.GetBlock(localX, localY, localZ)
}

// GetBoundingBox возвращает ограничивающий бокс чанка
func (c *Chunk) GetBoundingBox() physics.Box {
	return physics.Box{
		Min: c.Position,
		Max: c.Position.Add(mgl32.Vec3{
			ChunkWidth,
			ChunkHeight,
			ChunkWidth,
		}),
	}
}

// PositionToChunkCoords преобразует мировую позицию в координаты чанка
func PositionToChunkCoords(pos mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		float32(int(pos.X())/ChunkWidth) * ChunkWidth,
		0, // Чанки начинаются с Y=0
		float32(int(pos.Z())/ChunkWidth) * ChunkWidth,
	}
}

// GetChunkPosition возвращает позицию чанка
func (c *Chunk) GetChunkPosition() mgl32.Vec3 {
	return c.Position
}
