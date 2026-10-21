particle flame ~0.0 ~0.1 ~-1.5 0 0.5 0 0.4 0 normal
particle flame ~0.75 ~0.1 ~-1.299038 0 0.5 0 0.4 0 normal
particle flame ~1.299038 ~0.1 ~-0.75 0 0.5 0 0.4 0 normal
particle flame ~1.5 ~0.1 ~0.0 0 0.5 0 0.4 0 normal
particle flame ~1.299038 ~0.1 ~0.75 0 0.5 0 0.4 0 normal
particle flame ~0.75 ~0.1 ~1.299038 0 0.5 0 0.4 0 normal
particle flame ~-0.0 ~0.1 ~1.5 0 0.5 0 0.4 0 normal
particle flame ~-0.75 ~0.1 ~1.299038 0 0.5 0 0.4 0 normal
particle flame ~-1.299038 ~0.1 ~0.75 0 0.5 0 0.4 0 normal
particle flame ~-1.5 ~0.1 ~-0.0 0 0.5 0 0.4 0 normal
particle flame ~-1.299038 ~0.1 ~-0.75 0 0.5 0 0.4 0 normal
particle flame ~-0.75 ~0.1 ~-1.299038 0 0.5 0 0.4 0 normal

particle flame ~0.0 ~2 ~-1.5 0 -0.5 0 0.4 0 normal
particle flame ~0.75 ~2 ~-1.299038 0 -0.5 0 0.4 0 normal
particle flame ~1.299038 ~2 ~-0.75 0 -0.5 0 0.4 0 normal
particle flame ~1.5 ~2 ~0.0 0 -0.5 0 0.4 0 normal
particle flame ~1.299038 ~2 ~0.75 0 -0.5 0 0.4 0 normal
particle flame ~0.75 ~2 ~1.299038 0 -0.5 0 0.4 0 normal
particle flame ~-0.0 ~2 ~1.5 0 -0.5 0 0.4 0 normal
particle flame ~-0.75 ~2 ~1.299038 0 -0.5 0 0.4 0 normal
particle flame ~-1.299038 ~2 ~0.75 0 -0.5 0 0.4 0 normal
particle flame ~-1.5 ~2 ~-0.0 0 -0.5 0 0.4 0 normal
particle flame ~-1.299038 ~2 ~-0.75 0 -0.5 0 0.4 0 normal
particle flame ~-0.75 ~2 ~-1.299038 0 -0.5 0 0.4 0 normal

particle lava ~ ~1 ~ 0.3 0 0.3 1 5
particle fireworksSpark ~ ~1 ~ 0.3 0.3 0.3 0.4 4 normal

function rpg:monster/hurt/elemet/fire/main_fire

scoreboard players tag @s remove mob.target
effect @s minecraft:instant_damage 1 1
effect @s minecraft:instant_health 1 1