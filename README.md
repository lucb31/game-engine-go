# LIVE
[Try it out](https://lucb31.github.io/game-engine-go/)

# Survival game

## Gameplay


### Shop
- Shop only available if inside castle
- Add upgrades to buy with wood

### Harvesting
- Add animation / effect to identify which object is being harvested

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


# Useful commands
Generating release notes 
`git-cliff --unreleased --tag v0.3-alpha`
