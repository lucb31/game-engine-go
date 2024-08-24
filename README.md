# LIVE
[Try it out](https://lucb31.github.io/game-engine-go/)

# Survival game
## Parsing map data
- BUG: Cannot draw fence on exact outer bounds. need to keep 1 tile buffer. Dont know why that is

# Bugs
- Dash cooldown not working correctly
## Priority
- Add tutorial HUD menu
    - WASD to move
    - B to bring up shop
    - C to bring up player stats
    - D to bring up debugging menu

- Performance optimization for pathfinding algorithm
    - Quick win: Static graph

- eye frames during dash

## Open points
- Should the player be able to shoot from within the castle? Or pacified?

# Engine

## Damage model
- Add some randomization to demonstrate damage model (might remove / disable later) 
- Better armor model

## Projectiles
### Fix
- Put upper limit to nr of projectiles. Might cause problems otherwise

## Towers

### Feat
- Add targetting algorithm: Currently closest to tower 

