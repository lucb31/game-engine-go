package hud

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
)

type GameInfo interface {
	CreepProgress() ProgressInfo
	CastleProgress() ProgressInfo
	// Interface for speed slider to set game speed
	SetSpeed(float64)
	// Current money
	Balance() int64
	// Game over will be displayed if true
	GameOver() bool
	// Current score
	Score() ScoreValue
}

type SubMenu interface {
	Update()
	RootContainer() *widget.Container
}

type GameHUD struct {
	ui   *ebitenui.UI
	game GameInfo

	creepProgress     *widget.ProgressBar
	creepLabel        *widget.Text
	castleHealth      *widget.ProgressBar
	speedSlider       *widget.Slider
	goldLabel         *widget.Text
	gameOverContainer *widget.Container
	gameOverScore     *widget.Text

	scoreBoard ScoreBoard

	subMenus []SubMenu
}

func NewHUD(game GameInfo) (*GameHUD, error) {
	hud := &GameHUD{game: game}
	// Setup scoreboard: Use in memory in web env
	_, err := os.Getwd()
	if err != nil {
		hud.scoreBoard, err = NewInMemoryScoreBoard()
	} else {
		hud.scoreBoard, err = NewCsvScoreKeeper("data/score.csv")
	}
	if err != nil {
		return nil, err
	}

	// This creates the root container for this UI.
	// All other UI elements must be added to this container.
	rootContainer := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))

	hud.creepProgress, hud.creepLabel = initCreepProgressBar(rootContainer)
	hud.castleHealth = initCastleHealthProgressBar(rootContainer)
	hud.speedSlider = hud.initGameSpeedSlider(rootContainer)
	hud.goldLabel = initGoldLabel(rootContainer)
	hud.gameOverContainer = hud.initGameOverContainer(rootContainer)

	// This adds the root container to the UI, so that it will be rendered.
	hud.ui = &ebitenui.UI{
		Container: rootContainer,
	}

	return hud, nil
}

func (h *GameHUD) Draw(screen *ebiten.Image) {
	h.ui.Draw(screen)
}

func (h *GameHUD) Update() {
	h.ui.Update()
	h.updateCreepProgress()
	h.updateCastleHealth()
	h.updateGold()
	h.updateGameOver()

	// Draw submenus
	for _, menu := range h.subMenus {
		menu.Update()
	}
}

func (h *GameHUD) AddSubMenu(menu SubMenu) {
	h.subMenus = append(h.subMenus, menu)
	h.ui.Container.AddChild(menu.RootContainer())
}

func (h *GameHUD) SaveScore(score ScoreValue) {
	fmt.Printf("You've earned a score of %f\n", score)
	if h.scoreBoard.IsHighscore(score) {
		fmt.Println("NEW HIGHSCORE!")
	}
	err := h.scoreBoard.Save(score)
	if err != nil {
		fmt.Println("Could not save score", err.Error())
	}
	if err = h.scoreBoard.Print(); err != nil {
		fmt.Println("Could not print scoreboard", err.Error())
	}
}

func initCreepProgressBar(root *widget.Container) (*widget.ProgressBar, *widget.Text) {
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
	progress, _ := progressBarWithLabel(root, "Castle Health", layout, bgColor)
	return progress
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

func progressBarWithLabel(root *widget.Container, label string, anchor widget.AnchorLayoutData, bgColor color.Color) (*widget.ProgressBar, *widget.Text) {
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
	return progressBar, labelText
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

func (hud *GameHUD) initGameOverContainer(root *widget.Container) *widget.Container {
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(16),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	ttfFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal("Error Parsing Font", err)
	}
	// Game over label
	fontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: 32,
	})
	gameOverLabel := widget.NewText(
		widget.TextOpts.Text("Game Over", fontFace, color.RGBA{255, 0, 0, 1}),
	)
	container.AddChild(gameOverLabel)

	// Space to continue
	fontFace = truetype.NewFace(ttfFont, &truetype.Options{
		Size: 16,
	})
	restartLabel := widget.NewText(
		widget.TextOpts.Text("Press SPACE to restart", fontFace, color.RGBA{255, 255, 255, 1}),
	)
	container.AddChild(restartLabel)

	// Display final score
	fontFace = truetype.NewFace(ttfFont, &truetype.Options{
		Size: 16,
	})
	hud.gameOverScore = widget.NewText(
		widget.TextOpts.Text("Final score: 0", fontFace, color.RGBA{255, 255, 255, 1}),
	)
	container.AddChild(hud.gameOverScore)

	// Disable game over elements by default
	container.GetWidget().Visibility = widget.Visibility_Hide
	root.AddChild(container)
	return container
}

func (h *GameHUD) updateCreepProgress() {
	progress := h.game.CreepProgress()
	h.creepProgress.Min = progress.Min
	h.creepProgress.Max = progress.Max
	h.creepProgress.SetCurrent(progress.Current)
	h.creepLabel.Label = progress.Label
}

func (h *GameHUD) updateCastleHealth() {
	progress := h.game.CastleProgress()
	h.castleHealth.Min = progress.Min
	h.castleHealth.Max = progress.Max
	h.castleHealth.SetCurrent(progress.Current)
}

func (h *GameHUD) updateGold() {
	h.goldLabel.Label = fmt.Sprintf("Gold: %d", h.game.Balance())
}

func (h *GameHUD) updateGameOver() {
	// Toggle visibility of game over container
	if !h.game.GameOver() {
		h.gameOverContainer.GetWidget().Visibility = widget.Visibility_Hide
		return
	}
	h.gameOverContainer.GetWidget().Visibility = widget.Visibility_Show

	// Update final score
	currentScore := h.game.Score()
	var newScoreLabel string
	if h.scoreBoard.IsHighscore(currentScore) {
		newScoreLabel = fmt.Sprintf("HIGHSCORE: %1.1f", currentScore)
	} else {
		newScoreLabel = fmt.Sprintf("Final score: %1.1f (BEST %1.1f)", currentScore, h.scoreBoard.Highscore().Score)
	}
	h.gameOverScore.Label = newScoreLabel
}
