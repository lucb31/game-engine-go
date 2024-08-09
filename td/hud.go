package td

import (
	"image/color"
	"log"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
)

type GameHUD struct {
	ui   *ebitenui.UI
	game *TDGame

	creepProgress *widget.ProgressBar
}

type ProgressInfo struct {
	min     int
	max     int
	current int
}

func NewHUD(game *TDGame) (*GameHUD, error) {
	hud := &GameHUD{game: game}
	// This creates the root container for this UI.
	// All other UI elements must be added to this container.
	rootContainer := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))

	hud.creepProgress = initCreepProgressBar(rootContainer)

	// This adds the root container to the UI, so that it will be rendered.
	hud.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	return hud, nil
}

func initCreepProgressBar(root *widget.Container) *widget.ProgressBar {
	creepProgressContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		// Set the required anchor layout data to determine where in the root
		// container to place the progress bars.
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            widget.NewInsetsSimple(4),
			}),
		),
	)
	creepProgress := widget.NewProgressBar(
		widget.ProgressBarOpts.WidgetOpts(
			// Set the minimum size for the progress bar.
			// This is necessary if you wish to have the progress bar be larger than
			// the provided track image. In this exampe since we are using NineSliceColor
			// which is 1px x 1px we must set a minimum size.
			widget.WidgetOpts.MinSize(200, 16),
		),
		widget.ProgressBarOpts.Images(
			// Set the track images (Idle, Disabled).
			&widget.ProgressBarImage{
				Idle: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 150}),
			},
			// Set the progress images (Idle, Disabled).
			&widget.ProgressBarImage{
				Idle: image.NewNineSliceColor(color.NRGBA{0, 0, 255, 255}),
			},
		),
		// Set the min, max, and current values.
		widget.ProgressBarOpts.Values(0, 20, 20),
		// Set how much of the track is displayed when the bar is overlayed.
		widget.ProgressBarOpts.TrackPadding(widget.Insets{
			Top:    2,
			Bottom: 2,
		}),
	)
	creepProgressContainer.AddChild(creepProgress)
	root.AddChild(creepProgressContainer)

	// Init label
	ttfFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal("Error Parsing Font", err)
	}
	fontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: 16,
	})
	creepProgressLabel := widget.NewText(
		widget.TextOpts.Text("Wave 1", fontFace, color.White),
	)
	creepProgressContainer.AddChild(creepProgressLabel)
	return creepProgress
}

func (h *GameHUD) Draw(screen *ebiten.Image) {
	h.ui.Draw(screen)
}

func (h *GameHUD) Update() {
	h.ui.Update()
	h.updateCreepProgress()
}

func (h *GameHUD) updateCreepProgress() {
	progress := h.game.GetCreepProgress()
	h.creepProgress.Min = progress.min
	h.creepProgress.Max = progress.max
	h.creepProgress.SetCurrent(progress.current)
}
