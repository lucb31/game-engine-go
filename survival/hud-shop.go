package survival

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/lucb31/game-engine-go/engine"
	"golang.org/x/image/font/gofont/goregular"
)

type ShopMenu struct {
	// Dependencies
	ui          *ebitenui.UI
	goldManager engine.GoldManager
	playerStats engine.GameEntityStatReadWriter

	// UI
	shopContainer *widget.Container
	visible       bool

	// Logic
	itemSlots []*ShopItemSlot
}

type ShopItemSlot struct {
	item             *GameItem
	buyButton        *widget.Button
	priceLabel       *widget.Text
	descriptionLabel *widget.Text
}

// TODO: move struct to engine package
type GameItem struct {
	Price           int64
	Description     string
	ApplyItemEffect func(p engine.GameEntityStatReadWriter) error
}

// Pool of all available items in the shop
// TODO: Move to survival package
var availableItems = []GameItem{
	{Price: 50, Description: "+10 Max Health", ApplyItemEffect: func(p engine.GameEntityStatReadWriter) error {
		// Need to increase both max health & current health
		p.SetMaxHealth(p.MaxHealth() + 10.0)
		p.SetHealth(p.Health() + 10.0)
		return nil
	}},
	{Price: 50, Description: "+10 Movement speed", ApplyItemEffect: func(p engine.GameEntityStatReadWriter) error {
		p.SetMovementSpeed(p.MovementSpeed() + 10.0)
		return nil
	}},
	{Price: 50, Description: "+10 Armor", ApplyItemEffect: func(p engine.GameEntityStatReadWriter) error {
		p.SetArmor(p.Armor() + 10.0)
		return nil
	}},
	{Price: 50, Description: "+10 Power", ApplyItemEffect: func(p engine.GameEntityStatReadWriter) error {
		p.SetPower(p.Power() + 10.0)
		return nil
	}},
}

const itemSlots = 3

func NewShopMenu(goldManager engine.GoldManager, playerStats engine.GameEntityStatReadWriter) (*ShopMenu, error) {
	shop := &ShopMenu{goldManager: goldManager, playerStats: playerStats}
	return shop, nil
}

func (s *ShopMenu) RerollItemSlot(idx int) {
	// Select item from item pool
	itemIdx := rand.Intn(len(availableItems))
	newItem := &availableItems[itemIdx]

	// Update UI
	s.itemSlots[idx].item = newItem
	s.itemSlots[idx].priceLabel.Label = fmt.Sprintf("%d gold", newItem.Price)
	s.itemSlots[idx].descriptionLabel.Label = newItem.Description
}

func (s *ShopMenu) BuyHandler(idx int) {
	item := s.itemSlots[idx].item
	if !s.goldManager.CanAfford(item.Price) {
		fmt.Println("Cannot afford item", item)
		return
	}
	newBalance, err := s.goldManager.Remove(item.Price)
	if err != nil {
		fmt.Println("Error removing item cost", err.Error())
		return
	}
	err = item.ApplyItemEffect(s.playerStats)
	if err != nil {
		fmt.Println("Error applying item effect", err.Error())
		return
	}
	fmt.Printf("Bought item %v, new balance %d\n", item, newBalance)

	// Reroll
	s.RerollItemSlot(idx)
}

func (s *ShopMenu) Update() {
	// Toggle shop visibility with B
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		s.visible = !s.visible
	}
	if s.visible {
		s.shopContainer.GetWidget().Visibility = widget.Visibility_Show
	} else {
		s.shopContainer.GetWidget().Visibility = widget.Visibility_Hide
	}

	// Enable/disable buttons depending on affordability
	for _, slot := range s.itemSlots {
		if slot.buyButton == nil {
			continue
		}
		slot.buyButton.GetWidget().Disabled = !s.goldManager.CanAfford(slot.item.Price)
	}
}

func defaultApplyItemEffect(p engine.GameEntityStatReadWriter) error {
	return fmt.Errorf("Missing implementation")
}

func (s *ShopMenu) SetUI(ui *ebitenui.UI) {
	s.ui = ui
	// Init
	// construct a new container that serves as the root of the UI hierarchy
	rootContainer := widget.NewContainer(
		// the container will use a plain color as its background
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xbb})),
		// the container will use an anchor layout to layout its single child widget
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			//Define number of columns in the grid
			widget.GridLayoutOpts.Columns(3),
			//Define how much padding to inset the child content
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(30)),
			//Define how far apart the rows and columns should be
			widget.GridLayoutOpts.Spacing(20, 10),
			//Define how to stretch the rows and columns. Note it is required to
			//specify the Stretch for each row and column.
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{false, true}),
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
		fmt.Println("Error Parsing Font", err)
	}
	fontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: 16,
	})
	buttonImage, err := loadButtonImage()
	if err != nil {
		fmt.Println("Could not load button image", err.Error())
	}

	// Initialize item slots
	s.itemSlots = make([]*ShopItemSlot, itemSlots)
	for idx := range s.itemSlots {
		slot := &ShopItemSlot{}
		s.itemSlots[idx] = slot

		itemContainer := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{66, 66, 66, 255})),
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{true, true}),
			)),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(150, 150)),
		)

		slot.priceLabel = widget.NewText(
			widget.TextOpts.Text("", fontFace, color.RGBA{255, 255, 255, 1}),
			widget.TextOpts.MaxWidth(100),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart),
		)
		itemContainer.AddChild(slot.priceLabel)

		slot.descriptionLabel = widget.NewText(
			widget.TextOpts.Text("", fontFace, color.RGBA{255, 255, 255, 1}),
			widget.TextOpts.MaxWidth(100),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart),
		)
		itemContainer.AddChild(slot.descriptionLabel)

		slot.buyButton = widget.NewButton(
			widget.ButtonOpts.Image(buttonImage),
			widget.ButtonOpts.Text("Buy!", fontFace, &widget.ButtonTextColor{
				Idle: color.RGBA{255, 255, 255, 1},
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { s.BuyHandler(idx) }),
		)
		itemContainer.AddChild(slot.buyButton)
		rootContainer.AddChild(itemContainer)

		// Reroll to initialize
		s.RerollItemSlot(idx)
	}
	rootContainer.GetWidget().Visibility = widget.Visibility_Hide
	s.shopContainer = rootContainer
	s.ui.Container.AddChild(rootContainer)
}

func loadButtonImage() (*widget.ButtonImage, error) {
	idle := image.NewNineSliceColor(color.NRGBA{R: 0, G: 170, B: 0, A: 255})
	disabled := image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255})
	hover := image.NewNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255})
	pressed := image.NewNineSliceColor(color.NRGBA{R: 100, G: 100, B: 120, A: 255})

	return &widget.ButtonImage{
		Idle:     idle,
		Hover:    hover,
		Pressed:  pressed,
		Disabled: disabled,
	}, nil
}
