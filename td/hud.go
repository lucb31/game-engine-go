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
	castleHealth  *widget.ProgressBar
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
	hud.castleHealth = initCastleHealthProgressBar(rootContainer)

	// This adds the root container to the UI, so that it will be rendered.
	hud.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	return hud, nil
}

func initCreepProgressBar(root *widget.Container) *widget.ProgressBar {
	layout := widget.AnchorLayoutData{
		HorizontalPosition: widget.AnchorLayoutPositionCenter,
		VerticalPosition:   widget.AnchorLayoutPositionStart,
		Padding:            widget.NewInsetsSimple(4),
	}
	bgColor := color.NRGBA{0, 0, 255, 255}
	return progressBarWithLabel(root, "Wave 1", layout, bgColor)
}

func initCastleHealthProgressBar(root *widget.Container) *widget.ProgressBar {
	layout := widget.AnchorLayoutData{
		HorizontalPosition: widget.AnchorLayoutPositionEnd,
		VerticalPosition:   widget.AnchorLayoutPositionEnd,
		Padding:            widget.NewInsetsSimple(8),
	}
	bgColor := color.NRGBA{255, 0, 0, 255}
	return progressBarWithLabel(root, "Castle Health", layout, bgColor)
}

func progressBarWithLabel(root *widget.Container, label string, anchor widget.AnchorLayoutData, bgColor color.Color) *widget.ProgressBar {
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		// Set the required anchor layout data to determine where in the root
		// container to place the progress bars.
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(anchor),
		),
	)
	progressBar := widget.NewProgressBar(
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
				Idle: image.NewNineSliceColor(bgColor),
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
	container.AddChild(progressBar)
	root.AddChild(container)

	// Init label
	ttfFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal("Error Parsing Font", err)
	}
	fontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: 16,
	})
	labelText := widget.NewText(
		widget.TextOpts.Text(label, fontFace, color.White),
	)
	container.AddChild(labelText)
	return progressBar
}

func (h *GameHUD) Draw(screen *ebiten.Image) {
	h.ui.Draw(screen)
}

func (h *GameHUD) Update() {
	h.ui.Update()
	h.updateCreepProgress()
	h.updateCastleHealth()
}

func (h *GameHUD) updateCreepProgress() {
	progress := h.game.GetCreepProgress()
	h.creepProgress.Min = progress.min
	h.creepProgress.Max = progress.max
	h.creepProgress.SetCurrent(progress.current)
}

func (h *GameHUD) updateCastleHealth() {
	progress := h.game.GetCastleHealth()
	h.castleHealth.Min = progress.min
	h.castleHealth.Max = progress.max
	h.castleHealth.SetCurrent(progress.current)
}
