# LIVE
[Try it out](https://lucb31.github.io/game-engine-go/)

# Survival game

## Bug
- Sometimes day cycle is skipped and next wave is spawned immediately. Probably related to high game speeds

### Shop
- Add upgrades to buy with wood

### Harvesting
- Add animation / effect to identify which object is being harvested
- Different types of trees

## Priority
- Add tutorial HUD menu
    - WASD to move
    - B to bring up shop
    - E to interact
    - C to bring up player stats
    - D to bring up debugging menu

- Settings to control
    - Debug settings
    - Game speed

# Engine

## Damage model
- Applying damage should be handled by engine, not damage model. Otherwise there will bo too many cross-dependencies
- Add some randomization to demonstrate damage model (might remove / disable later) 
- Better armor model

## Projectiles
### Fix
- Put upper limit to nr of projectiles. Might cause problems otherwise


# Useful commands
Generating release notes 
`git-cliff --unreleased --tag v0.3-alpha`
