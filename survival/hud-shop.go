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
	"github.com/lucb31/game-engine-go/engine/loot"
	"golang.org/x/image/font/gofont/goregular"
)

type ShopMenu struct {
	// Dependencies
	ui          *ebitenui.UI
	inventory   loot.Inventory
	playerStats engine.GameEntityStatReadWriter

	// UI
	shopContainer *widget.Container
	visible       bool

	// Logic
	itemSlots []*ShopItemSlot
}

type ShopItemSlot struct {
	item             *loot.GameItem
	buyButton        *widget.Button
	rerollButton     *widget.Button
	priceLabel       *widget.Text
	descriptionLabel *widget.Text
}

func (i *ShopItemSlot) ApplyItemEffect(tar engine.GameEntityStatReadWriter) error {
	effect, ok := itemEffects[i.item.ItemEffectId]
	if !ok {
		return fmt.Errorf("Unknown item efect id: %d", i.item.ItemEffectId)
	}
	return effect(tar)
}

type ItemEffect func(p engine.GameEntityStatReadWriter) error

var itemEffects = map[loot.ItemEffectId]ItemEffect{
	loot.ItemEffectAddMaxHealth: func(p engine.GameEntityStatReadWriter) error {
		// Need to increase both max health & current health
		p.SetMaxHealth(p.MaxHealth() + 10.0)
		p.SetHealth(p.Health() + 10.0)
		return nil
	},
	loot.ItemEffectAddMovementSpeed: func(p engine.GameEntityStatReadWriter) error {
		p.SetMovementSpeed(p.MovementSpeed() + 10.0)
		return nil
	},
	loot.ItemEffectAddArmor: func(p engine.GameEntityStatReadWriter) error {
		p.SetArmor(p.Armor() + 10.0)
		return nil
	},
	loot.ItemEffectAddPower: func(p engine.GameEntityStatReadWriter) error {
		p.SetPower(p.Power() + 10.0)
		return nil
	},
	loot.ItemEffectAddAtkSpeed: func(p engine.GameEntityStatReadWriter) error {
		p.SetAtkSpeed(p.AtkSpeed() + 0.2)
		return nil
	},
}

// Pool of all available items in the shop
// TODO: Move static list to survival package
var availableItems = []loot.GameItem{
	{Price: 50, Description: "+10 Max Health", ItemEffectId: loot.ItemEffectAddMaxHealth},
	{Price: 50, Description: "+10 Movement speed", ItemEffectId: loot.ItemEffectAddMovementSpeed},
	{Price: 50, Description: "+10 Armor", ItemEffectId: loot.ItemEffectAddArmor},
	{Price: 50, Description: "+10 Power", ItemEffectId: loot.ItemEffectAddPower},
	{Price: 50, Description: "+0.2 Atk Speed", ItemEffectId: loot.ItemEffectAddAtkSpeed},
}

const (
	itemSlots   = 3
	rerollPrice = 10
)

func NewShopMenu(inventory loot.Inventory, playerStats engine.GameEntityStatReadWriter) (*ShopMenu, error) {
	shop := &ShopMenu{inventory: inventory, playerStats: playerStats}
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
	shopItem := s.itemSlots[idx]
	gameItem := shopItem.item
	if !s.inventory.CanAfford(gameItem.Price) {
		fmt.Println("Cannot afford item", gameItem)
		return
	}
	newBalance, err := s.inventory.Spend(gameItem.Price)
	if err != nil {
		fmt.Println("Error removing item cost", err.Error())
		return
	}

	err = shopItem.ApplyItemEffect(s.playerStats)
	if err != nil {
		fmt.Println("Error applying item effect", err.Error())
		return
	}
	fmt.Printf("Bought item %v, new balance %d\n", gameItem, newBalance)

	s.RerollItemSlot(idx)
}

func (s *ShopMenu) RerollHandler(idx int) {
	if !s.inventory.CanAfford(rerollPrice) {
		fmt.Println("Cannot afford to reroll")
		return
	}
	newBalance, err := s.inventory.Spend(rerollPrice)
	if err != nil {
		fmt.Println("Error removing item cost", err.Error())
		return
	}
	fmt.Printf("Rerolled item slot %d, new balance %d\n", idx, newBalance)

	s.RerollItemSlot(idx)
}
func (s *ShopMenu) Update() {
	// Toggle shop visibility with B
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		s.visible = !s.visible
	}
	// Close shop with ESC
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.visible = false
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
		slot.buyButton.GetWidget().Disabled = !s.inventory.CanAfford(slot.item.Price)
		if slot.rerollButton == nil {
			continue
		}
		slot.rerollButton.GetWidget().Disabled = !s.inventory.CanAfford(rerollPrice)
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
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false, true, false}),
				widget.GridLayoutOpts.Spacing(0, 10),
			)),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(150, 150)),
		)

		slot.rerollButton = widget.NewButton(
			widget.ButtonOpts.Image(buttonImage),
			widget.ButtonOpts.Text(fmt.Sprintf("Reroll (%dg)", rerollPrice), fontFace, &widget.ButtonTextColor{
				Idle: color.RGBA{255, 255, 255, 1},
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { s.RerollHandler(idx) }),
		)
		itemContainer.AddChild(slot.rerollButton)

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
