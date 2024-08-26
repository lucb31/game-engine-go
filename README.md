# LIVE
[Try it out](https://lucb31.github.io/game-engine-go/)

# Survival game

## Parsing map data
- Automatically draw outer bounds walls

- BUG: Cannot draw fence on exact outer bounds. need to keep 1 tile buffer. Dont know why that is

## Gameplay

### Castle
- Castle only shoots if player inside (maybe using the player gun?)
- Remove shooting from player
- Improve bounding box

### Shop
- Shop only available if inside castle
- Add upgrades to buy with wood


## Priority
- Add tutorial HUD menu
    - WASD to move
    - B to bring up shop
    - E to interact
    - C to bring up player stats
    - D to bring up debugging menu

- eye frames during dash

## Open points
- Should the player be able to shoot from within the castle? Or pacified?

# Engine

## Damage model
- Applying damage should be handled by engine, not damage model. Otherwise there will bo too many cross-dependencies
- Add some randomization to demonstrate damage model (might remove / disable later) 
- Better armor model

## Projectiles
### Fix
- Put upper limit to nr of projectiles. Might cause problems otherwise


