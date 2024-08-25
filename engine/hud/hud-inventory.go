package hud

import (
	"fmt"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/lucb31/game-engine-go/engine/loot"
	"golang.org/x/image/font/gofont/goregular"
)

type InventoryHud struct {
	rootContainer *widget.Container
	inventory     loot.Inventory
	goldLabel     *widget.Text
	woodLabel     *widget.Text
}

func NewInventoryHud(inventory loot.Inventory) (*InventoryHud, error) {
	hud := &InventoryHud{inventory: inventory}
	hud.rootContainer = hud.buildRootContainer()

	hud.goldLabel = hud.buildLabel()
	hud.rootContainer.AddChild(hud.goldLabel)

	hud.woodLabel = hud.buildLabel()
	hud.rootContainer.AddChild(hud.woodLabel)

	return hud, nil
}

func (h *InventoryHud) Update() {
	h.updateGold()
	h.updateWood()
}

func (h *InventoryHud) RootContainer() *widget.Container { return h.rootContainer }

func (h *InventoryHud) buildRootContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, false}, []bool{false}),
			widget.GridLayoutOpts.Spacing(10, 0),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				Padding:            widget.Insets{Bottom: 30, Right: 8},
			}),
		),
	)
}

func (h *InventoryHud) buildLabel() *widget.Text {
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
	return labelText
}

func (h *InventoryHud) updateGold() {
	h.goldLabel.Label = fmt.Sprintf("Gold: %d", h.inventory.GoldManager().Balance())
}

func (h *InventoryHud) updateWood() {
	h.woodLabel.Label = fmt.Sprintf("Wood: %d", h.inventory.WoodManager().Balance())
}
