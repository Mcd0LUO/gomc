particle flame ~0.0 ~0.8 ~-3.0 -0.0 0 1.0 0.5 0 normal
particle flame ~1.5 ~0.8 ~-2.598076 -0.5 0 0.866025 0.5 0 normal
particle flame ~2.598076 ~0.8 ~-1.5 -0.866025 0 0.5 0.5 0 normal
particle flame ~3.0 ~0.8 ~0.0 -1.0 0 -0.0 0.5 0 normal
particle flame ~2.598076 ~0.8 ~1.5 -0.866025 0 -0.5 0.5 0 normal
particle flame ~1.5 ~0.8 ~2.598076 -0.5 0 -0.866025 0.5 0 normal
particle flame ~-0.0 ~0.8 ~3.0 0.0 0 -1.0 0.5 0 normal
particle flame ~-1.5 ~0.8 ~2.598076 0.5 0 -0.866025 0.5 0 normal
particle flame ~-2.598076 ~0.8 ~1.5 0.866025 0 -0.5 0.5 0 normal
particle flame ~-3.0 ~0.8 ~-0.0 1.0 0 0.0 0.5 0 normal
particle flame ~-2.598076 ~0.8 ~-1.5 0.866025 0 0.5 0.5 0 normal
particle flame ~-1.5 ~0.8 ~-2.598076 0.5 0 0.866025 0.5 0 normal
particle lava ~ ~1 ~ 0.3 0 0.3 1 5
particle fireworksSpark ~ ~1 ~ 0.3 0.3 0.3 0.4 4 normal

function rpg:monster/hurt/elemet/fire/main_fire

scoreboard players tag @s remove mob.target
effect @s minecraft:instant_damage 1 1
effect @s minecraft:instant_health 1 1