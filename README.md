<div align="center">
    <img alt="Gopher" src="https://github.com/golang-samples/gopher-vector/blob/master/gopher.png" height="250" />
    <br />
    <h3>Go Game Engine</h3>
    <p align="center">
        <a href="https://lucb31.github.io/game-engine-go/">Online Demo</a>
    </p>
</div>

# About
Summer project 2024. Goals
 - Learn fundamentals of go
 - Understand and solve common Game engines problems
 - Apply game engine to survival / tower defense themed game

## Built with
- [Ebit Engine](https://github.com/hajimehoshi/ebiten)

## Demo
### Windows
Download & run latest `.exe` artifact from [release section](https://github.com/lucb31/game-engine-go/releases). 

### Browser version
Although not optimized for browser support, the application can be run as a WebGL application available [here](https://lucb31.github.io/game-engine-go/). A stable frame rate cannot be guaranteed.


# Feature coverage: Engine
- [x] Camera movement & zoom
- [x] Proceduaral level generation 
- [x] Multi-layered tile map rendering
- [x] Character control
- [x] Animation management
- [x] SFX management
- [x] Shop, loot, resource management system
- [x] Damage model
- [x] Fog of war

----

# Backlog Survival game
## Features

### Damage model
- Add some randomization to demonstrate damage model (might remove / disable later) 
- More sophisticated armor model

### Upgrade system
- Piercing projectiles upgrade
- Splitting projectiles upgrade
- Upgrade tree hierarchy / path

### Harvesting
- Add animation / effect to identify which object is being harvested
- Add randomization in loot system / only drop wood from specific trees 

### UI
- Add tutorial HUD menu
    - WASD to move
    - B to bring up shop
    - E to interact
    - C to bring up player stats
    - D to bring up debugging menu

- Settings to control
    - Debug settings
    - Game speed

## Bugs
- Improve sync of creep swing animation & SFX
- Improve visual feedback of day & night cycle


## Refactoring & improvement ideas
### Damage model
- Applying damage should be handled by engine, not damage model. Otherwise there will bo too many cross-dependencies

### Projectiles
- Put upper limit to nr of projectiles. Might cause problems otherwise

# Useful commands
Generating release notes 
`git-cliff --unreleased --tag v0.3-alpha`
