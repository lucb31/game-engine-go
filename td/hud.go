package td

import (
	"fmt"
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
	speedSlider   *widget.Slider
	goldLabel     *widget.Text
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
	hud.speedSlider = hud.initGameSpeedSlider(rootContainer)
	hud.goldLabel = initGoldLabel(rootContainer)

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

func (h *GameHUD) initGameSpeedSlider(root *widget.Container) *widget.Slider {
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout()),
		// Set the required anchor layout data to determine where in the root
		// container to place the progress bars.
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				Padding:            widget.Insets{Bottom: 30, Right: 8},
			}),
		),
	)

	// construct a slider (ebitenui/examples/slider/main.go)
	slider := widget.NewSlider(
		// Set the slider orientation - n/s vs e/w
		widget.SliderOpts.Direction(widget.DirectionHorizontal),
		// Set the minimum and maximum value for the slider
		widget.SliderOpts.MinMax(0, 100),

		widget.SliderOpts.WidgetOpts(
			// Set the Widget to layout in the center on the screen
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
			}),
			// Set the widget's dimensions
			widget.WidgetOpts.MinSize(200, 6),
		),
		widget.SliderOpts.Images(
			// Set the track images
			&widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
			},
			// Set the handle images
			&widget.ButtonImage{
				Idle:    image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Hover:   image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
				Pressed: image.NewNineSliceColor(color.NRGBA{255, 100, 100, 255}),
			},
		),
		// Set the size of the handle
		widget.SliderOpts.FixedHandleSize(6),
		// Set the offset to display the track
		widget.SliderOpts.TrackOffset(0),
		// Set the size to move the handle
		widget.SliderOpts.PageSizeFunc(func() int {
			return 1
		}),
		// Set the callback to call when the slider value is changed
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			h.game.SetSpeed(float64(args.Current) * 0.1)
		}),
	)
	slider.Current = 10
	container.AddChild(slider)
	root.AddChild(container)
	return slider
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

func initGoldLabel(root *widget.Container) *widget.Text {
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				Padding:            widget.Insets{Bottom: 30, Right: 8},
			}),
		),
	)

	// Init label
	ttfFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal("Error Parsing Font", err)
	}
	fontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: 16,
	})
	labelText := widget.NewText(
		widget.TextOpts.Text("", fontFace, color.RGBA{244, 228, 0, 1}),
	)
	container.AddChild(labelText)
	root.AddChild(container)
	return labelText
}

func (h *GameHUD) Draw(screen *ebiten.Image) {
	h.ui.Draw(screen)
}

func (h *GameHUD) Update() {
	h.ui.Update()
	h.updateCreepProgress()
	h.updateCastleHealth()
	h.updateGold()
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

func (h *GameHUD) updateGold() {
	h.goldLabel.Label = fmt.Sprintf("Gold: %d", h.game.Balance())
}
