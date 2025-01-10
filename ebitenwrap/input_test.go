package ebitenwrap

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockKeyboard struct {
	pressed map[ebiten.Key]bool
}

func newMockKeyboard() *mockKeyboard {
	return &mockKeyboard{
		pressed: make(map[ebiten.Key]bool),
	}
}

func (k *mockKeyboard) IsKeyPressed(key ebiten.Key) bool {
	return k.pressed[key]
}

type mockMouse struct {
	pressed map[ebiten.MouseButton]bool
}

func newMockMouse() *mockMouse {
	return &mockMouse{
		pressed: make(map[ebiten.MouseButton]bool),
	}
}

func (m *mockMouse) IsMouseButtonPressed(button ebiten.MouseButton) bool {
	return m.pressed[button]
}

type mockTouch struct {
	activeIDs []ebiten.TouchID
	positions map[ebiten.TouchID][2]int
}

func newMockTouch() *mockTouch {
	return &mockTouch{
		positions: make(map[ebiten.TouchID][2]int),
	}
}

func (t *mockTouch) AppendTouchIDs(touches []ebiten.TouchID) []ebiten.TouchID {
	return append(touches, t.activeIDs...)
}

func (t *mockTouch) TouchPosition(id ebiten.TouchID) (int, int) {
	pos := t.positions[id]
	return pos[0], pos[1]
}

func newTestInputManager(t testing.TB) *InputManager {
	keyboard := newMockKeyboard()
	mouse := newMockMouse()
	touch := newMockTouch()

	manager, err := NewInputManager(keyboard, mouse, touch)
	require.NoError(t, err)

	return manager
}

func TestKeyboardInput(t *testing.T) {
	manager := newTestInputManager(t)
	keyboard := manager.keyboard.(*mockKeyboard)

	// Test key press detection
	t.Run("key press detection", func(t *testing.T) {
		key := ebiten.KeyA

		// Key not pressed initially
		assert.False(t, manager.IsKeyJustPressed(key))
		assert.Equal(t, 0, manager.KeyPressDuration(key))

		// Press key
		keyboard.pressed[key] = true
		manager.Tick()
		assert.True(t, manager.IsKeyJustPressed(key))
		assert.Equal(t, 1, manager.KeyPressDuration(key))

		// Hold key
		manager.Tick()
		assert.False(t, manager.IsKeyJustPressed(key))
		assert.Equal(t, 2, manager.KeyPressDuration(key))

		// Release key
		keyboard.pressed[key] = false
		manager.Tick()
		assert.False(t, manager.IsKeyJustPressed(key))
		assert.Equal(t, 0, manager.KeyPressDuration(key))
		assert.True(t, manager.IsKeyJustReleased(key))
	})
}

func TestMouseInput(t *testing.T) {
	manager := newTestInputManager(t)
	mouse := manager.mouse.(*mockMouse)

	t.Run("mouse button press detection", func(t *testing.T) {
		button := ebiten.MouseButtonLeft

		// Button not pressed initially
		assert.False(t, manager.IsMouseButtonJustPressed(button))
		assert.Equal(t, 0, manager.MouseButtonPressDuration(button))

		// Press button
		mouse.pressed[button] = true
		manager.Tick()
		assert.True(t, manager.IsMouseButtonJustPressed(button))
		assert.Equal(t, 1, manager.MouseButtonPressDuration(button))

		// Hold button
		manager.Tick()
		assert.False(t, manager.IsMouseButtonJustPressed(button))
		assert.Equal(t, 2, manager.MouseButtonPressDuration(button))

		// Release button
		mouse.pressed[button] = false
		manager.Tick()
		assert.False(t, manager.IsMouseButtonJustPressed(button))
		assert.Equal(t, 0, manager.MouseButtonPressDuration(button))
		assert.True(t, manager.IsMouseButtonJustReleased(button))
	})
}

func TestTouchInput(t *testing.T) {
	manager := newTestInputManager(t)
	touch := manager.touch.(*mockTouch)

	t.Run("touch detection and position", func(t *testing.T) {
		touchID := ebiten.TouchID(1)
		expectedPos := [2]int{100, 200}

		// No touch initially
		var touchIDs []ebiten.TouchID
		touchIDs = manager.AppendJustPressedTouchIDs(touchIDs)
		assert.Empty(t, touchIDs)
		assert.Equal(t, 0, manager.TouchPressDuration(touchID))

		// Start touch
		touch.activeIDs = []ebiten.TouchID{touchID}
		touch.positions[touchID] = expectedPos
		manager.Tick()

		// Verify touch position
		x, y := manager.TouchPosition(touchID)
		assert.Equal(t, expectedPos[0], x)
		assert.Equal(t, expectedPos[1], y)

		// Verify touch press detection
		touchIDs = manager.AppendJustPressedTouchIDs(touchIDs[:0])
		assert.Len(t, touchIDs, 1)
		assert.Equal(t, touchID, touchIDs[0])
		assert.Equal(t, 1, manager.TouchPressDuration(touchID))

		// Hold touch
		manager.Tick()
		touchIDs = manager.AppendJustPressedTouchIDs(touchIDs[:0])
		assert.Empty(t, touchIDs)
		assert.Equal(t, 2, manager.TouchPressDuration(touchID))

		// Release touch
		touch.activeIDs = nil
		manager.Tick()

		// Verify touch release
		touchIDs = manager.AppendJustReleasedTouchIDs(touchIDs[:0])
		assert.Len(t, touchIDs, 1)
		assert.Equal(t, touchID, touchIDs[0])
		assert.True(t, manager.IsTouchJustReleased(touchID))
	})
}

func TestInputStateInterfaces(t *testing.T) {
	manager := newTestInputManager(t)

	t.Run("input state interface implementation", func(t *testing.T) {
		var inputState InputState = manager
		assert.NotNil(t, inputState.Keyboard())
		assert.NotNil(t, inputState.Mouse())
		assert.NotNil(t, inputState.Gamepad())
		assert.NotNil(t, inputState.Touch())
	})

	t.Run("keyboard state interface implementation", func(t *testing.T) {
		var keyboardState KeyboardState = manager
		var keys []ebiten.Key
		keys = keyboardState.AppendPressedKeys(make([]ebiten.Key, 0))
		assert.NotNil(t, keys)
		keys = keyboardState.AppendJustPressedKeys(make([]ebiten.Key, 0))
		assert.NotNil(t, keys)
		keys = keyboardState.AppendJustReleasedKeys(make([]ebiten.Key, 0))
		assert.NotNil(t, keys)
	})

	t.Run("mouse state interface implementation", func(t *testing.T) {
		var mouseState MouseState = manager
		assert.NotPanics(t, func() {
			mouseState.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
			mouseState.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
			mouseState.MouseButtonPressDuration(ebiten.MouseButtonLeft)
		})
	})

	t.Run("touch state interface implementation", func(t *testing.T) {
		var touchState TouchState = manager
		var touchIDs []ebiten.TouchID
		touchIDs = touchState.AppendJustPressedTouchIDs(make([]ebiten.TouchID, 0))
		assert.NotNil(t, touchIDs)
		touchIDs = touchState.AppendJustReleasedTouchIDs(make([]ebiten.TouchID, 0))
		assert.NotNil(t, touchIDs)
	})
}
