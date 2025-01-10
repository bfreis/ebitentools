package ebitenwrap

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Tick struct {
	InputState InputState
	TPS        int
}

type Game interface {
	Update(tick Tick) error
	Draw(screen *ebiten.Image)
	Layout(outsideWidth int, outsideHeight int) (screenWidth, screenHeight int)
}

type Wrapper struct {
	game         Game
	inputManager *InputManager
}

var _ ebiten.Game = (*Wrapper)(nil)

type WrapOptions struct {
	Keyboard Keyboard
	Mouse    Mouse
	Touch    Touch
}

func Wrap(game Game, inputManager *InputManager) (*Wrapper, error) {
	return &Wrapper{
		game:         game,
		inputManager: inputManager,
	}, nil
}

func (w *Wrapper) Update() error {
	w.inputManager.Tick()

	t := Tick{
		InputState: w.inputManager,
		TPS:        ebiten.TPS(),
	}

	return w.game.Update(t)
}

func (w *Wrapper) Draw(screen *ebiten.Image) {
	w.game.Draw(screen)
}

func (w *Wrapper) Layout(outsideWidth, outsideHeight int) (int, int) {
	return w.game.Layout(outsideWidth, outsideHeight)
}
