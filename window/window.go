package window

import (
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Config содержит настройки окна
type Config struct {
	Width        int
	Height       int
	Title        string
	Resizable    bool
	CaptureMouse bool
}

// DefaultConfig возвращает конфигурацию окна по умолчанию
func DefaultConfig() Config {
	return Config{
		Width:        1280,
		Height:       720,
		Title:        "Game Engine",
		Resizable:    false,
		CaptureMouse: true,
	}
}

// Window представляет собой обертку над glfw.Window с дополнительной функциональностью
type Window struct {
	window   *glfw.Window
	debounce map[glfw.Key]bool
	config   Config
}

// New создает новое окно с заданной конфигурацией
func New(config Config) (*Window, error) {
	runtime.LockOSThread()
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	if config.Resizable {
		glfw.WindowHint(glfw.Resizable, glfw.True)
	} else {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	}

	window, err := glfw.CreateWindow(config.Width, config.Height, config.Title, nil, nil)
	if err != nil {
		glfw.Terminate()
		return nil, err
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		window.Destroy()
		glfw.Terminate()
		return nil, err
	}

	if config.CaptureMouse {
		window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	}

	w := &Window{
		window:   window,
		debounce: make(map[glfw.Key]bool),
		config:   config,
	}

	return w, nil
}

// Terminate закрывает окно и освобождает ресурсы GLFW
func (w *Window) Terminate() {
	w.window.Destroy()
	glfw.Terminate()
}

// ShouldClose проверяет, должно ли окно закрыться
func (w *Window) ShouldClose() bool {
	return w.window.ShouldClose()
}

// Update обновляет состояние окна
func (w *Window) Update() {
	w.window.SwapBuffers()
	glfw.PollEvents()
}

// IsPressed возвращает true, если клавиша нажата
func (w *Window) IsPressed(k glfw.Key) bool {
	return w.window.GetKey(k) == glfw.Press
}

// IsReleased возвращает true, если клавиша отпущена
func (w *Window) IsReleased(k glfw.Key) bool {
	return w.window.GetKey(k) == glfw.Release
}

// Debounce предотвращает повторные срабатывания клавиши и возвращает true, если клавиша была нажата впервые
func (w *Window) Debounce(k glfw.Key) bool {
	debounce := w.debounce[k]
	if w.IsPressed(k) && !debounce {
		w.debounce[k] = true
		return true
	} else if w.IsReleased(k) {
		delete(w.debounce, k)
	}
	return false
}

// SetCursorMode управляет режимом курсора
func (w *Window) SetCursorMode(mode int) {
	w.window.SetInputMode(glfw.CursorMode, mode)
}

// GetConfig возвращает текущую конфигурацию окна
func (w *Window) GetConfig() Config {
	return w.config
}

// SetCursorPosCallback устанавливает колбэк для отслеживания позиции курсора
func (w *Window) SetCursorPosCallback(callback glfw.CursorPosCallback) {
	w.window.SetCursorPosCallback(callback)
}

// GetGLFWWindow возвращает нативный glfw.Window
func (w *Window) GetGLFWWindow() *glfw.Window {
	return w.window
}
