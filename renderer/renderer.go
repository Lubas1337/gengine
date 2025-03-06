package renderer

import (
	"fmt"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/user/gengine/physics"
	"github.com/user/gengine/world"
)

// Структура для элемента управления
type ControlInfo struct {
	Key  string
	Desc string
}

// Renderer представляет рендерер для отрисовки игровых объектов
type Renderer struct {
	shader     uint32
	vao, vbo   uint32
	projection mgl32.Mat4
	view       mgl32.Mat4

	// Для подсчета FPS
	frameCount  int
	lastFpsTime time.Time
	currentFps  int
}

// NewRenderer создает новый рендерер
func NewRenderer(width, height int) (*Renderer, error) {
	r := &Renderer{
		lastFpsTime: time.Now(),
		frameCount:  0,
		currentFps:  0,
	}

	// Настраиваем OpenGL для видимости всех сторон
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Disable(gl.CULL_FACE) // Отключаем отсечение граней для видимости всех сторон

	// Включаем блендинг
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Создаем простой шейдер
	vertexShaderSource := `
		#version 410
		layout (location = 0) in vec3 position;
		layout (location = 1) in vec3 color;
		
		uniform mat4 projection;
		uniform mat4 view;
		uniform mat4 model;
		
		out vec3 fragColor;
		
		void main() {
			gl_Position = projection * view * model * vec4(position, 1.0);
			fragColor = color;
		}
	` + "\x00"

	fragmentShaderSource := `
		#version 410
		in vec3 fragColor;
		out vec4 color;
		
		void main() {
			color = vec4(fragColor, 1.0);
		}
	` + "\x00"

	// Компилируем шейдеры
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	// Создаем программу
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Проверяем ошибки линковки
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength)
		gl.GetProgramInfoLog(program, logLength, nil, &log[0])

		return nil, fmt.Errorf("Ошибка линковки шейдерной программы: %s", string(log))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	r.shader = program

	// Создаем буферы
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)

	// Настраиваем атрибуты
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 6*4, 0)

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, 6*4, 3*4)

	// Настраиваем матрицы проекции и вида
	aspect := float32(width) / float32(height)
	r.projection = mgl32.Perspective(mgl32.DegToRad(70.0), aspect, 0.1, 1000.0) // Увеличиваем угол обзора
	r.view = mgl32.LookAtV(
		mgl32.Vec3{0, 0, 10},
		mgl32.Vec3{0, 0, 0},
		mgl32.Vec3{0, 1, 0},
	)

	return r, nil
}

// SetCamera устанавливает позицию и направление камеры
func (r *Renderer) SetCamera(position, target, up mgl32.Vec3) {
	r.view = mgl32.LookAtV(position, target, up)
}

// Begin начинает рендеринг кадра
func (r *Renderer) Begin() {
	// Очищаем буферы
	gl.ClearColor(0.1, 0.1, 0.1, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Включаем нужные функции
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Используем шейдерную программу
	gl.UseProgram(r.shader)

	// Привязываем VAO
	gl.BindVertexArray(r.vao)

	// Устанавливаем матрицы проекции и вида
	projLoc := gl.GetUniformLocation(r.shader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &r.projection[0])

	viewLoc := gl.GetUniformLocation(r.shader, gl.Str("view\x00"))
	gl.UniformMatrix4fv(viewLoc, 1, false, &r.view[0])
}

// End завершает рендеринг кадра
func (r *Renderer) End() {
	// Увеличиваем счетчик кадров
	r.frameCount++

	// Проверяем, прошла ли секунда для обновления FPS
	now := time.Now()
	elapsed := now.Sub(r.lastFpsTime)
	if elapsed.Seconds() >= 1.0 {
		r.currentFps = int(float64(r.frameCount) / elapsed.Seconds())
		r.frameCount = 0
		r.lastFpsTime = now
	}
}

// DrawChunk отрисовывает чанк
func (r *Renderer) DrawChunk(chunk *world.Chunk) {
	// Проверяем наличие чанка
	if chunk == nil {
		return
	}

	// Отрисовываем каждый блок отдельно для надежности
	for x := 0; x < world.ChunkWidth; x++ {
		for y := 0; y < world.ChunkHeight; y++ {
			for z := 0; z < world.ChunkWidth; z++ {
				block := chunk.GetBlock(x, y, z)
				if block != nil && block.Active {
					// Создаем матрицу модели для блока
					modelLoc := gl.GetUniformLocation(r.shader, gl.Str("model\x00"))
					blockPos := block.Position
					blockModel := mgl32.Translate3D(blockPos.X(), blockPos.Y(), blockPos.Z()).Mul4(
						mgl32.Scale3D(0.98, 0.98, 0.98)) // Чуть меньше 1, чтобы были видны грани
					gl.UniformMatrix4fv(modelLoc, 1, false, &blockModel[0])

					// Выбираем цвет в зависимости от типа блока
					var color mgl32.Vec3
					switch block.BlockType {
					case "stone":
						color = mgl32.Vec3{0.5, 0.5, 0.5} // Серый для камня
					case "brick":
						color = mgl32.Vec3{0.8, 0.2, 0.2} // Красный для кирпича
					default:
						color = mgl32.Vec3{0.3, 0.3, 0.8} // Синий для остальных
					}

					// Рисуем блок
					r.drawSolidCube(color)
				}
			}
		}
	}
}

// drawBlockBatch рисует группу блоков одного типа для оптимизации
func (r *Renderer) drawBlockBatch(positions []mgl32.Vec3, color mgl32.Vec3) {
	// Если позиций нет, ничего не делаем
	if len(positions) == 0 {
		return
	}

	// Для очень большого количества блоков используем оптимизированный подход
	if len(positions) > 100 {
		// Просто выборочно отрисовываем часть блоков для повышения производительности
		// В реальном рендерере здесь мог бы быть инстансинг
		step := len(positions)/100 + 1
		for i := 0; i < len(positions); i += step {
			modelLoc := gl.GetUniformLocation(r.shader, gl.Str("model\x00"))
			blockPos := positions[i]
			blockModel := mgl32.Translate3D(blockPos.X(), blockPos.Y(), blockPos.Z()).Mul4(
				mgl32.Scale3D(0.98, 0.98, 0.98))
			gl.UniformMatrix4fv(modelLoc, 1, false, &blockModel[0])
			r.drawSolidCube(color)
		}
	} else {
		// Отрисовываем каждый блок
		for _, pos := range positions {
			modelLoc := gl.GetUniformLocation(r.shader, gl.Str("model\x00"))
			blockModel := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z()).Mul4(
				mgl32.Scale3D(0.98, 0.98, 0.98))
			gl.UniformMatrix4fv(modelLoc, 1, false, &blockModel[0])
			r.drawSolidCube(color)
		}
	}
}

// drawSolidCube рисует заполненный куб с заданным цветом
func (r *Renderer) drawSolidCube(color mgl32.Vec3) {
	// Упрощенная версия вершин для куба (36 вершин - по 3 на треугольник, по 2 треугольника на грань, 6 граней)
	vertices := []float32{
		// Позиции и цвета вершин
		// Передняя грань (z = 0.5)
		-0.5, -0.5, 0.5, color.X(), color.Y(), color.Z(), // левый нижний
		0.5, -0.5, 0.5, color.X(), color.Y(), color.Z(), // правый нижний
		0.5, 0.5, 0.5, color.X(), color.Y(), color.Z(), // правый верхний
		0.5, 0.5, 0.5, color.X(), color.Y(), color.Z(), // правый верхний
		-0.5, 0.5, 0.5, color.X(), color.Y(), color.Z(), // левый верхний
		-0.5, -0.5, 0.5, color.X(), color.Y(), color.Z(), // левый нижний

		// Задняя грань (z = -0.5)
		-0.5, -0.5, -0.5, color.X() * 0.8, color.Y() * 0.8, color.Z() * 0.8, // левый нижний
		-0.5, 0.5, -0.5, color.X() * 0.8, color.Y() * 0.8, color.Z() * 0.8, // левый верхний
		0.5, 0.5, -0.5, color.X() * 0.8, color.Y() * 0.8, color.Z() * 0.8, // правый верхний
		0.5, 0.5, -0.5, color.X() * 0.8, color.Y() * 0.8, color.Z() * 0.8, // правый верхний
		0.5, -0.5, -0.5, color.X() * 0.8, color.Y() * 0.8, color.Z() * 0.8, // правый нижний
		-0.5, -0.5, -0.5, color.X() * 0.8, color.Y() * 0.8, color.Z() * 0.8, // левый нижний

		// Левая грань (x = -0.5)
		-0.5, -0.5, -0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // левый нижний зад
		-0.5, -0.5, 0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // левый нижний перед
		-0.5, 0.5, 0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // левый верхний перед
		-0.5, 0.5, 0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // левый верхний перед
		-0.5, 0.5, -0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // левый верхний зад
		-0.5, -0.5, -0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // левый нижний зад

		// Правая грань (x = 0.5)
		0.5, -0.5, -0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // правый нижний зад
		0.5, 0.5, -0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // правый верхний зад
		0.5, 0.5, 0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // правый верхний перед
		0.5, 0.5, 0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // правый верхний перед
		0.5, -0.5, 0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // правый нижний перед
		0.5, -0.5, -0.5, color.X() * 0.7, color.Y() * 0.7, color.Z() * 0.7, // правый нижний зад

		// Верхняя грань (y = 0.5)
		-0.5, 0.5, -0.5, color.X() * 0.9, color.Y() * 0.9, color.Z() * 0.9, // левый верхний зад
		-0.5, 0.5, 0.5, color.X() * 0.9, color.Y() * 0.9, color.Z() * 0.9, // левый верхний перед
		0.5, 0.5, 0.5, color.X() * 0.9, color.Y() * 0.9, color.Z() * 0.9, // правый верхний перед
		0.5, 0.5, 0.5, color.X() * 0.9, color.Y() * 0.9, color.Z() * 0.9, // правый верхний перед
		0.5, 0.5, -0.5, color.X() * 0.9, color.Y() * 0.9, color.Z() * 0.9, // правый верхний зад
		-0.5, 0.5, -0.5, color.X() * 0.9, color.Y() * 0.9, color.Z() * 0.9, // левый верхний зад

		// Нижняя грань (y = -0.5)
		-0.5, -0.5, -0.5, color.X() * 0.6, color.Y() * 0.6, color.Z() * 0.6, // левый нижний зад
		0.5, -0.5, -0.5, color.X() * 0.6, color.Y() * 0.6, color.Z() * 0.6, // правый нижний зад
		0.5, -0.5, 0.5, color.X() * 0.6, color.Y() * 0.6, color.Z() * 0.6, // правый нижний перед
		0.5, -0.5, 0.5, color.X() * 0.6, color.Y() * 0.6, color.Z() * 0.6, // правый нижний перед
		-0.5, -0.5, 0.5, color.X() * 0.6, color.Y() * 0.6, color.Z() * 0.6, // левый нижний перед
		-0.5, -0.5, -0.5, color.X() * 0.6, color.Y() * 0.6, color.Z() * 0.6, // левый нижний зад
	}

	// Передаем данные в GPU
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Включаем атрибуты вершин
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 6*4, 0)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, 6*4, 3*4)

	// Рисуем треугольники (36 вершин = 12 треугольников = 6 граней куба)
	gl.DrawArrays(gl.TRIANGLES, 0, 36)
}

// DrawBox отрисовывает коллайдер
func (r *Renderer) DrawBox(box physics.Box, color mgl32.Vec3) {
	modelLoc := gl.GetUniformLocation(r.shader, gl.Str("model\x00"))

	// Вычисляем размеры и центр бокса
	size := box.Max.Sub(box.Min)
	center := box.Min.Add(size.Mul(0.5))

	// Создаем матрицу модели
	model := mgl32.Translate3D(center.X(), center.Y(), center.Z()).Mul4(
		mgl32.Scale3D(size.X(), size.Y(), size.Z()))

	gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

	// Рисуем каркас коллайдера
	r.drawWireframe(color)
}

// drawWireframe рисует каркас куба с заданным цветом
func (r *Renderer) drawWireframe(color mgl32.Vec3) {
	// Вершины для каркаса куба
	vertices := []float32{
		// Позиция           // Цвет
		-0.5, -0.5, -0.5, color.X(), color.Y(), color.Z(),
		0.5, -0.5, -0.5, color.X(), color.Y(), color.Z(),
		0.5, 0.5, -0.5, color.X(), color.Y(), color.Z(),
		-0.5, 0.5, -0.5, color.X(), color.Y(), color.Z(),
		-0.5, -0.5, 0.5, color.X(), color.Y(), color.Z(),
		0.5, -0.5, 0.5, color.X(), color.Y(), color.Z(),
		0.5, 0.5, 0.5, color.X(), color.Y(), color.Z(),
		-0.5, 0.5, 0.5, color.X(), color.Y(), color.Z(),
	}

	// Индексы для рисования линий
	indices := []uint32{
		0, 1, 1, 2, 2, 3, 3, 0, // Нижняя грань
		4, 5, 5, 6, 6, 7, 7, 4, // Верхняя грань
		0, 4, 1, 5, 2, 6, 3, 7, // Соединяющие линии
	}

	// Передаем данные в GPU
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Создаем временный индексный буфер
	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Рисуем линии
	gl.DrawElements(gl.LINES, int32(len(indices)), gl.UNSIGNED_INT, nil)

	// Освобождаем индексный буфер
	gl.DeleteBuffers(1, &ebo)
}

// DrawControls отрисовывает таблицу с управлением
func (r *Renderer) DrawControls(controls []struct{ Key, Desc string }) {
	// В простом случае просто выводим в консоль
	if len(controls) > 0 {
		fmt.Println("=== Управление ===")
		for _, control := range controls {
			fmt.Printf("[%s]: %s\n", control.Key, control.Desc)
		}
		fmt.Println("=================")
	}

	// В реальной реализации здесь был бы код для отрисовки текста или UI на экране
	// Но так как это требует дополнительные ресурсы (текстуры шрифтов, текстовый рендерер),
	// ограничимся выводом в консоль
}

// Destroy освобождает ресурсы рендерера
func (r *Renderer) Destroy() {
	gl.DeleteProgram(r.shader)
	gl.DeleteBuffers(1, &r.vbo)
	gl.DeleteVertexArrays(1, &r.vao)
}

// compileShader компилирует шейдер и возвращает его идентификатор
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	// Проверяем ошибки компиляции
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])

		return 0, fmt.Errorf("Ошибка компиляции шейдера: %s", string(log))
	}

	return shader, nil
}

// GetFPS возвращает текущее значение FPS
func (r *Renderer) GetFPS() int {
	return r.currentFps
}

// DrawFPS отрисовывает значение FPS на экране
func (r *Renderer) DrawFPS() {
	fmt.Printf("FPS: %d\n", r.currentFps)
	// В реальной реализации здесь был бы код для отрисовки текста на экране
}
