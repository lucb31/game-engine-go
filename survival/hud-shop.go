package survival

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"strings"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/lucb31/game-engine-go/engine"
	"github.com/lucb31/game-engine-go/engine/loot"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

type ShopEnabler interface {
	ShopEnabled() bool
}
type GunProvider interface {
	Gun() engine.Gun
}

type ShopMenu struct {
	// Dependencies
	inventory   loot.Inventory
	playerStats engine.GameEntityStatReadWriter
	// Gun that stat upgrades will be applied to
	GunProvider

	// UI
	shopContainer *widget.Container
	// Visibility toggled by B key
	visible bool
	// Enabled status toggled by entering & leaving the castle
	ShopEnabler

	// Logic
	randomItemSlots []*ShopItemSlot
	fixedItemSlots  []*ShopItemSlot
}

type ShopItemSlot struct {
	item             *loot.GameItem
	buyButton        *widget.Button
	rerollButton     *widget.Button
	priceLabel       *widget.Text
	descriptionLabel *widget.Text
}

type ItemEffectContext struct {
	engine.GameEntityStatReadWriter
	gun engine.Gun
}

func (i *ShopItemSlot) ApplyItemEffect(ctx *ItemEffectContext) error {
	effect, ok := itemEffects[i.item.ItemEffectId]
	if !ok {
		return fmt.Errorf("Unknown item efect id: %d", i.item.ItemEffectId)
	}
	return effect(ctx)
}

func (i *ShopItemSlot) generatePriceLabel() string {
	prices := []string{}
	if i.item.GoldPrice > 0 {
		prices = append(prices, fmt.Sprintf("%d gold", i.item.GoldPrice))
	}
	if i.item.WoodPrice > 0 {
		prices = append(prices, fmt.Sprintf("%d wood", i.item.WoodPrice))
	}
	return strings.Join(prices, ",")
}

type ItemEffect func(*ItemEffectContext) error

var itemEffects = map[loot.ItemEffectId]ItemEffect{
	loot.ItemEffectAddMaxHealth: func(p *ItemEffectContext) error {
		// Need to increase both max health & current health
		p.SetMaxHealth(p.MaxHealth() + 10.0)
		p.SetHealth(p.Health() + 10.0)
		return nil
	},
	loot.ItemEffectAddMovementSpeed: func(p *ItemEffectContext) error {
		p.SetMovementSpeed(p.MovementSpeed() + 10.0)
		return nil
	},
	loot.ItemEffectAddArmor: func(p *ItemEffectContext) error {
		p.SetArmor(p.Armor() + 10.0)
		return nil
	},
	loot.ItemEffectAddPower: func(p *ItemEffectContext) error {
		p.SetPower(p.Power() + 10.0)
		return nil
	},
	loot.ItemEffectAddAtkSpeed: func(p *ItemEffectContext) error {
		p.SetAtkSpeed(p.AtkSpeed() + 0.2)
		return nil
	},
	loot.ItemEffectAddCastleProjectile: func(ctx *ItemEffectContext) error {
		if ctx.gun == nil {
			return fmt.Errorf("Could not add projectile: No gun provided")
		}
		ctx.gun.SetProjectileCount(ctx.gun.ProjectileCount() + 1)
		return nil
	},
}

// Pool of all available items in the shop. X items from this pool will be randomly selected
var availableItems = []loot.GameItem{
	{GoldPrice: 50, Description: "+10 Max Health", ItemEffectId: loot.ItemEffectAddMaxHealth},
	{GoldPrice: 50, Description: "+10 Movement speed", ItemEffectId: loot.ItemEffectAddMovementSpeed},
	{GoldPrice: 50, Description: "+10 Armor", ItemEffectId: loot.ItemEffectAddArmor},
	{GoldPrice: 50, Description: "+10 Power", ItemEffectId: loot.ItemEffectAddPower},
	{GoldPrice: 50, Description: "+0.2 Atk Speed", ItemEffectId: loot.ItemEffectAddAtkSpeed},
}

// Pool of permanent upgrades. All will be available
var fixedUpgrades = []loot.GameItem{
	{WoodPrice: 50, Description: "Additional projectile", ItemEffectId: loot.ItemEffectAddCastleProjectile},
}

const (
	randomizedItemSlots = 3
	rerollPrice         = 10
)

func NewShopMenu(inventory loot.Inventory, playerStats engine.GameEntityStatReadWriter) (*ShopMenu, error) {
	shop := &ShopMenu{inventory: inventory, playerStats: playerStats}
	shop.init()
	return shop, nil
}

func (s *ShopMenu) RerollItemSlot(idx int) {
	// Select item from item pool
	itemIdx := rand.Intn(len(availableItems))
	newItem := &availableItems[itemIdx]

	// Update UI
	s.randomItemSlots[idx].item = newItem
	s.randomItemSlots[idx].priceLabel.Label = s.randomItemSlots[idx].generatePriceLabel()
	s.randomItemSlots[idx].descriptionLabel.Label = newItem.Description
}

func (s *ShopMenu) BuyAndApply(shopItem *ShopItemSlot) error {
	gameItem := shopItem.item

	// Buy item via inventory (manages resources)
	err := s.inventory.Buy(gameItem)
	if err != nil {
		return fmt.Errorf("Could not buy game item: %s", err.Error())
	}

	// Apply item effect
	// TODO: Consider moving this somewhere else
	ctx := &ItemEffectContext{s.playerStats, s.Gun()}
	err = shopItem.ApplyItemEffect(ctx)
	if err != nil {
		return fmt.Errorf("Error applying item effect: %s", err.Error())
	}
	log.Printf("Successfully bought item %v\n", gameItem)
	return nil
}

func (s *ShopMenu) randomizedItemSlotBuyHandler(idx int) {
	shopItem := s.randomItemSlots[idx]
	if err := s.BuyAndApply(shopItem); err != nil {
		log.Println(err.Error())
	}
	// Reroll a new item into the slot
	s.RerollItemSlot(idx)
}

func (s *ShopMenu) RerollHandler(idx int) {
	if !s.inventory.GoldManager().CanAfford(rerollPrice) {
		log.Println("Cannot afford to reroll")
		return
	}
	newBalance, err := s.inventory.GoldManager().Remove(rerollPrice)
	if err != nil {
		log.Println("Error removing item cost", err.Error())
		return
	}
	log.Printf("Rerolled item slot %d, new balance %d\n", idx, newBalance)

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

	// Early exit: If not visible, we dont need to update anything
	if !s.visible {
		s.shopContainer.GetWidget().Visibility = widget.Visibility_Hide
		return
	}

	s.shopContainer.GetWidget().Visibility = widget.Visibility_Show

	// Enable/disable buttons depending on affordability & shop status
	rerollDisabled := !s.ShopEnabled() || !s.inventory.GoldManager().CanAfford(rerollPrice)
	shopDisabled := !s.ShopEnabled()
	for _, slot := range s.randomItemSlots {
		if slot.buyButton == nil {
			continue
		}
		canAffordItem, _ := s.inventory.CanAfford(slot.item)
		buyDisabled := shopDisabled || !canAffordItem
		slot.buyButton.GetWidget().Disabled = buyDisabled
		if slot.rerollButton == nil {
			continue
		}
		slot.rerollButton.GetWidget().Disabled = rerollDisabled
	}
	for _, slot := range s.fixedItemSlots {
		if slot.buyButton == nil {
			continue
		}
		canAffordItem, _ := s.inventory.CanAfford(slot.item)
		buyDisabled := shopDisabled || !canAffordItem
		slot.buyButton.GetWidget().Disabled = buyDisabled
	}
}

func (s *ShopMenu) init() {
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
	rootContainer.GetWidget().Visibility = widget.Visibility_Hide
	s.shopContainer = rootContainer

	// Load fonts & button assets
	ttfFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Println("Error Parsing Font", err)
	}
	fontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: 16,
	})
	buttonImage, err := loadButtonImage()
	if err != nil {
		log.Println("Could not load button image", err.Error())
	}

	// Init shop elements
	s.initRandomizedItemSlots(fontFace, buttonImage)
	s.initPermanantUpgradeItemSlots(fontFace, buttonImage)
}

func (s *ShopMenu) initRandomizedItemSlots(fontFace font.Face, buttonImage *widget.ButtonImage) {
	s.randomItemSlots = make([]*ShopItemSlot, randomizedItemSlots)
	for idx := range s.randomItemSlots {
		slot := &ShopItemSlot{}
		s.randomItemSlots[idx] = slot

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
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) { s.randomizedItemSlotBuyHandler(idx) }),
		)
		itemContainer.AddChild(slot.buyButton)

		s.shopContainer.AddChild(itemContainer)

		// Reroll to initialize
		s.RerollItemSlot(idx)
	}
}

func (s *ShopMenu) initPermanantUpgradeItemSlots(fontFace font.Face, buttonImage *widget.ButtonImage) {
	s.fixedItemSlots = make([]*ShopItemSlot, len(fixedUpgrades))
	for idx, item := range fixedUpgrades {
		slot := &ShopItemSlot{item: &item}
		s.fixedItemSlots[idx] = slot

		itemContainer := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{66, 66, 66, 255})),
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(1),
				widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, false, true, false}),
				widget.GridLayoutOpts.Spacing(0, 10),
			)),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(150, 150)),
		)

		slot.priceLabel = widget.NewText(
			widget.TextOpts.Text(slot.generatePriceLabel(), fontFace, color.RGBA{255, 255, 255, 1}),
			widget.TextOpts.MaxWidth(100),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart),
		)
		itemContainer.AddChild(slot.priceLabel)

		slot.descriptionLabel = widget.NewText(
			widget.TextOpts.Text(item.Description, fontFace, color.RGBA{255, 255, 255, 1}),
			widget.TextOpts.MaxWidth(100),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionStart),
		)
		itemContainer.AddChild(slot.descriptionLabel)

		slot.buyButton = widget.NewButton(
			widget.ButtonOpts.Image(buttonImage),
			widget.ButtonOpts.Text("Buy!", fontFace, &widget.ButtonTextColor{
				Idle: color.RGBA{255, 255, 255, 1},
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				if err := s.BuyAndApply(s.fixedItemSlots[idx]); err != nil {
					log.Println("Could not buy: ", err.Error())
				}
			}),
		)
		itemContainer.AddChild(slot.buyButton)
		s.shopContainer.AddChild(itemContainer)
	}
}

func (s *ShopMenu) RootContainer() *widget.Container   { return s.shopContainer }
func (s *ShopMenu) SetShopEnabler(enabler ShopEnabler) { s.ShopEnabler = enabler }
func (s *ShopMenu) SetGunProvider(gun GunProvider) {
	s.GunProvider = gun
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
