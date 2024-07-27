package engine

type GameObj struct {
	asset           *GameAsset
	activeAnimation string
	posX            int
	posY            int
}

type GameAsset struct {
	Name        string
	Animations  map[string]GameAssetAnimation
	frameHeight int8
	frameWidht  int8
}

type GameAssetAnimation struct {
	Name string
	// Number of frames to play
	frameCount int
	// Vertical position in the TileMap (will be multiplied by frameHeight)
	yPos int8
}
