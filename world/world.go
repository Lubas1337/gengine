package world

import (
	"fmt"
	"sync"

	"github.com/go-gl/mathgl/mgl32"
)

// World представляет собой мир, состоящий из чанков
type World struct {
	chunks      map[string]*Chunk
	chunksMutex sync.RWMutex
}

// NewWorld создает новый мир
func NewWorld() *World {
	return &World{
		chunks: make(map[string]*Chunk),
	}
}

// GetChunkKey генерирует ключ для чанка по его позиции
func GetChunkKey(pos mgl32.Vec3) string {
	return fmt.Sprintf("%d_%d", int(pos.X()), int(pos.Z()))
}

// AddChunk добавляет чанк в мир
func (w *World) AddChunk(chunk *Chunk) {
	w.chunksMutex.Lock()
	defer w.chunksMutex.Unlock()

	key := GetChunkKey(chunk.Position)
	w.chunks[key] = chunk
}

// GetChunk возвращает чанк по его позиции
func (w *World) GetChunk(pos mgl32.Vec3) *Chunk {
	w.chunksMutex.RLock()
	defer w.chunksMutex.RUnlock()

	chunkPos := PositionToChunkCoords(pos)
	key := GetChunkKey(chunkPos)
	return w.chunks[key]
}

// GetBlock возвращает блок по мировым координатам
func (w *World) GetBlock(pos mgl32.Vec3) *BlockData {
	chunk := w.GetChunk(pos)
	if chunk == nil {
		return nil
	}

	return chunk.GetBlockFromWorldPos(pos)
}

// SetBlock устанавливает блок по мировым координатам
func (w *World) SetBlock(pos mgl32.Vec3, blockType string, active bool) {
	chunk := w.GetChunk(pos)
	if chunk == nil {
		// Если чанк не существует, создаем его
		chunkPos := PositionToChunkCoords(pos)
		chunk = NewChunk(chunkPos)
		w.AddChunk(chunk)
	}

	// Вычисляем локальные координаты блока внутри чанка
	localX := int(pos.X()) % ChunkWidth
	localY := int(pos.Y())
	localZ := int(pos.Z()) % ChunkWidth

	// Обрабатываем отрицательные координаты
	if localX < 0 {
		localX += ChunkWidth
	}
	if localZ < 0 {
		localZ += ChunkWidth
	}

	chunk.SetBlock(localX, localY, localZ, blockType, active)
}

// GetAllChunks возвращает все чанки мира
func (w *World) GetAllChunks() []*Chunk {
	w.chunksMutex.RLock()
	defer w.chunksMutex.RUnlock()

	chunks := make([]*Chunk, 0, len(w.chunks))
	for _, chunk := range w.chunks {
		chunks = append(chunks, chunk)
	}

	return chunks
}

// GetChunksInRadius возвращает все чанки в заданном радиусе от точки
func (w *World) GetChunksInRadius(center mgl32.Vec3, radius float32) []*Chunk {
	w.chunksMutex.RLock()
	defer w.chunksMutex.RUnlock()

	radiusSq := radius * radius
	chunks := make([]*Chunk, 0)

	for _, chunk := range w.chunks {
		// Вычисляем центр чанка
		chunkCenter := chunk.Position.Add(mgl32.Vec3{
			ChunkWidth / 2,
			ChunkHeight / 2,
			ChunkWidth / 2,
		})

		// Проверяем, находится ли чанк в радиусе
		diff := chunkCenter.Sub(center)
		distSq := diff.X()*diff.X() + diff.Y()*diff.Y() + diff.Z()*diff.Z()
		if distSq <= radiusSq {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}
