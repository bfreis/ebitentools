package ebitenwrapfx

import (
	"github.com/bfreis/ebitentools/ebitenwrap"

	"go.uber.org/fx"
)

var Module = fx.Module("ebitenwrap",
	fx.Provide(ebitenwrap.NewInputManager),
	fx.Provide(ebitenwrap.Wrap),

	fx.Supply(fx.Annotate(&ebitenwrap.DefaultKeyboard{}, fx.As(new(ebitenwrap.Keyboard)))),
	fx.Supply(fx.Annotate(&ebitenwrap.DefaultMouse{}, fx.As(new(ebitenwrap.Mouse)))),
	fx.Supply(fx.Annotate(&ebitenwrap.DefaultTouch{}, fx.As(new(ebitenwrap.Touch)))),
)
