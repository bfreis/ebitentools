package ebitenwrap

import (
	"slices"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type InputManager struct {
	keyboard Keyboard
	mouse    Mouse
	touch    Touch

	mu sync.RWMutex

	keyDurations     [ebiten.KeyMax + 1]int
	prevKeyDurations [ebiten.KeyMax + 1]int

	mouseButtonDurations     [ebiten.MouseButtonMax + 1]int
	prevMouseButtonDurations [ebiten.MouseButtonMax + 1]int

	touchIDsBuf     []ebiten.TouchID
	touchStates     map[ebiten.TouchID]touchState
	prevTouchStates map[ebiten.TouchID]touchState
}

type touchState struct {
	duration int
	x        int
	y        int
}

func NewDefaultInputManager() (*InputManager, error) {
	return NewInputManager(DefaultKeyboard{}, DefaultMouse{}, DefaultTouch{})
}

func NewInputManager(
	keyboard Keyboard,
	mouse Mouse,
	touch Touch,
) (*InputManager, error) {
	return &InputManager{
		keyboard: keyboard,
		mouse:    mouse,
		touch:    touch,

		touchStates:     make(map[ebiten.TouchID]touchState),
		prevTouchStates: make(map[ebiten.TouchID]touchState),
	}, nil
}

type Keyboard interface {
	IsKeyPressed(key ebiten.Key) bool
}

type DefaultKeyboard struct{}

func (_ DefaultKeyboard) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

type Mouse interface {
	IsMouseButtonPressed(button ebiten.MouseButton) bool
}

type DefaultMouse struct{}

func (_ DefaultMouse) IsMouseButtonPressed(button ebiten.MouseButton) bool {
	return ebiten.IsMouseButtonPressed(button)
}

type Touch interface {
	AppendTouchIDs(touches []ebiten.TouchID) []ebiten.TouchID
	TouchPosition(id ebiten.TouchID) (int, int)
}

type DefaultTouch struct{}

func (_ DefaultTouch) AppendTouchIDs(touches []ebiten.TouchID) []ebiten.TouchID {
	return ebiten.AppendTouchIDs(touches)
}

func (_ DefaultTouch) TouchPosition(id ebiten.TouchID) (int, int) {
	return ebiten.TouchPosition(id)
}

func (m *InputManager) Tick() {
	m.mu.Lock()
	defer m.mu.Unlock()

	copy(m.prevKeyDurations[:], m.keyDurations[:])
	for idx := range m.keyDurations {
		if m.keyboard.IsKeyPressed(ebiten.Key(idx)) {
			m.keyDurations[idx]++
		} else {
			m.keyDurations[idx] = 0
		}
	}

	copy(m.prevMouseButtonDurations[:], m.mouseButtonDurations[:])
	for idx := range m.mouseButtonDurations {
		if m.mouse.IsMouseButtonPressed(ebiten.MouseButton(idx)) {
			m.mouseButtonDurations[idx]++
		} else {
			m.mouseButtonDurations[idx] = 0
		}
	}

	clear(m.prevTouchStates)
	for id, state := range m.touchStates {
		m.prevTouchStates[id] = state
	}

	m.touchIDsBuf = m.touch.AppendTouchIDs(m.touchIDsBuf[:0])
	for _, id := range m.touchIDsBuf {
		state := m.touchStates[id]
		state.duration++
		state.x, state.y = m.touch.TouchPosition(id)
		m.touchStates[id] = state
	}

	// Remove released touches.
	for id := range m.touchStates {
		if !slices.Contains(m.touchIDsBuf, id) {
			delete(m.touchStates, id)
		}
	}

}

func (m *InputManager) Keyboard() KeyboardState {
	return m
}

func (m *InputManager) AppendPressedKeys(keys []ebiten.Key) []ebiten.Key {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for key, ticks := range m.keyDurations {
		if ticks > 0 {
			keys = append(keys, ebiten.Key(key))
		}
	}

	return keys
}

func (m *InputManager) AppendJustPressedKeys(keys []ebiten.Key) []ebiten.Key {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for key, ticks := range m.keyDurations {
		if ticks == 1 {
			keys = append(keys, ebiten.Key(key))
		}
	}

	return keys
}

func (m *InputManager) AppendJustReleasedKeys(keys []ebiten.Key) []ebiten.Key {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for key := range m.keyDurations {
		if m.prevKeyDurations[key] > 0 && m.keyDurations[key] == 0 {
			keys = append(keys, ebiten.Key(key))
		}
	}

	return keys
}

func (m *InputManager) IsKeyJustPressed(key ebiten.Key) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.keyDurations[key] == 1
}

func (m *InputManager) IsKeyJustReleased(key ebiten.Key) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.prevKeyDurations[key] > 0 && m.keyDurations[key] == 0
}

func (m *InputManager) KeyPressDuration(key ebiten.Key) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.keyDurations[key]
}

func (m *InputManager) Mouse() MouseState {
	return m
}

func (m *InputManager) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.mouseButtonDurations[button] == 1
}

func (m *InputManager) IsMouseButtonJustReleased(button ebiten.MouseButton) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.prevMouseButtonDurations[button] > 0 && m.mouseButtonDurations[button] == 0
}

func (m *InputManager) MouseButtonPressDuration(button ebiten.MouseButton) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.mouseButtonDurations[button]
}

func (m *InputManager) Gamepad() GamepadState {
	return &DefaultGamepadState{}
}

func (m *InputManager) Touch() TouchState {
	return m
}

func (m *InputManager) AppendJustPressedTouchIDs(touchIDs []ebiten.TouchID) []ebiten.TouchID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	origLen := len(touchIDs)

	for id, ts := range m.touchStates {
		if ts.duration == 1 {
			touchIDs = append(touchIDs, id)
		}
	}

	slices.Sort(touchIDs[origLen:])

	return touchIDs
}

func (m *InputManager) AppendJustReleasedTouchIDs(touchIDs []ebiten.TouchID) []ebiten.TouchID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	origLen := len(touchIDs)

	for id, pts := range m.prevTouchStates {
		if pts.duration > 0 && m.touchStates[id].duration == 0 {
			touchIDs = append(touchIDs, id)
		}
	}

	slices.Sort(touchIDs[origLen:])

	return touchIDs
}

func (m *InputManager) IsTouchJustReleased(id ebiten.TouchID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.prevTouchStates[id].duration > 0 && m.touchStates[id].duration == 0
}

func (m *InputManager) TouchPressDuration(id ebiten.TouchID) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.touchStates[id].duration
}

func (m *InputManager) TouchPosition(id ebiten.TouchID) (int, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ts, ok := m.touchStates[id]; ok {
		return ts.x, ts.y
	}

	if pts, ok := m.prevTouchStates[id]; ok {
		return pts.x, pts.y
	}

	return 0, 0
}

type InputState interface {
	Keyboard() KeyboardState
	Mouse() MouseState
	Gamepad() GamepadState
	Touch() TouchState
}

type KeyboardState interface {
	AppendPressedKeys(keys []ebiten.Key) []ebiten.Key
	AppendJustPressedKeys(keys []ebiten.Key) []ebiten.Key
	AppendJustReleasedKeys(keys []ebiten.Key) []ebiten.Key
	IsKeyJustPressed(key ebiten.Key) bool
	IsKeyJustReleased(key ebiten.Key) bool
	KeyPressDuration(key ebiten.Key) int
}

type MouseState interface {
	IsMouseButtonJustPressed(button ebiten.MouseButton) bool
	IsMouseButtonJustReleased(button ebiten.MouseButton) bool
	MouseButtonPressDuration(button ebiten.MouseButton) int
}

type GamepadState interface {
	AppendJustConnectedGamepadIDs(gamepadIDs []ebiten.GamepadID) []ebiten.GamepadID
	IsGamepadJustDisconnected(id ebiten.GamepadID) bool
	AppendPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton
	AppendJustPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton
	AppendJustReleasedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton
	IsGamepadButtonJustPressed(id ebiten.GamepadID, button ebiten.GamepadButton) bool
	IsGamepadButtonJustReleased(id ebiten.GamepadID, button ebiten.GamepadButton) bool
	GamepadButtonPressDuration(id ebiten.GamepadID, button ebiten.GamepadButton) int
	AppendPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton
	AppendJustPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton
	AppendJustReleasedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton
	IsStandardGamepadButtonJustPressed(id ebiten.GamepadID, button ebiten.StandardGamepadButton) bool
	IsStandardGamepadButtonJustReleased(id ebiten.GamepadID, button ebiten.StandardGamepadButton) bool
	StandardGamepadButtonPressDuration(id ebiten.GamepadID, button ebiten.StandardGamepadButton) int
}

type TouchState interface {
	AppendJustPressedTouchIDs(touchIDs []ebiten.TouchID) []ebiten.TouchID
	AppendJustReleasedTouchIDs(touchIDs []ebiten.TouchID) []ebiten.TouchID
	IsTouchJustReleased(id ebiten.TouchID) bool
	TouchPressDuration(id ebiten.TouchID) int
	TouchPosition(id ebiten.TouchID) (int, int)
}

type DefaultGamepadState struct{}

func (_ DefaultGamepadState) AppendJustConnectedGamepadIDs(gamepadIDs []ebiten.GamepadID) []ebiten.GamepadID {
	return inpututil.AppendJustConnectedGamepadIDs(gamepadIDs)
}

func (_ DefaultGamepadState) IsGamepadJustDisconnected(id ebiten.GamepadID) bool {
	return inpututil.IsGamepadJustDisconnected(id)
}

func (_ DefaultGamepadState) AppendPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton {
	return inpututil.AppendPressedGamepadButtons(id, buttons)
}

func (_ DefaultGamepadState) AppendJustPressedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton {
	return inpututil.AppendJustPressedGamepadButtons(id, buttons)
}

func (_ DefaultGamepadState) AppendJustReleasedGamepadButtons(id ebiten.GamepadID, buttons []ebiten.GamepadButton) []ebiten.GamepadButton {
	return inpututil.AppendJustReleasedGamepadButtons(id, buttons)
}

func (_ DefaultGamepadState) IsGamepadButtonJustPressed(id ebiten.GamepadID, button ebiten.GamepadButton) bool {
	return inpututil.IsGamepadButtonJustPressed(id, button)
}

func (_ DefaultGamepadState) IsGamepadButtonJustReleased(id ebiten.GamepadID, button ebiten.GamepadButton) bool {
	return inpututil.IsGamepadButtonJustReleased(id, button)
}

func (_ DefaultGamepadState) GamepadButtonPressDuration(id ebiten.GamepadID, button ebiten.GamepadButton) int {
	return inpututil.GamepadButtonPressDuration(id, button)
}

func (_ DefaultGamepadState) AppendPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
	return inpututil.AppendPressedStandardGamepadButtons(id, buttons)
}

func (_ DefaultGamepadState) AppendJustPressedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
	return inpututil.AppendJustPressedStandardGamepadButtons(id, buttons)
}

func (_ DefaultGamepadState) AppendJustReleasedStandardGamepadButtons(id ebiten.GamepadID, buttons []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
	return inpututil.AppendJustReleasedStandardGamepadButtons(id, buttons)
}

func (_ DefaultGamepadState) IsStandardGamepadButtonJustPressed(id ebiten.GamepadID, button ebiten.StandardGamepadButton) bool {
	return inpututil.IsStandardGamepadButtonJustPressed(id, button)
}

func (_ DefaultGamepadState) IsStandardGamepadButtonJustReleased(id ebiten.GamepadID, button ebiten.StandardGamepadButton) bool {
	return inpututil.IsStandardGamepadButtonJustReleased(id, button)
}

func (_ DefaultGamepadState) StandardGamepadButtonPressDuration(id ebiten.GamepadID, button ebiten.StandardGamepadButton) int {
	return inpututil.StandardGamepadButtonPressDuration(id, button)
}
