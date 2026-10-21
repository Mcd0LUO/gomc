particle smoke ~ ~ ~ 0 0 0 0.25 20 normal
particle witchMagic ~ ~2.2 ~ 0 0 0 0.25 10 normal
particle endRod ~ ~1 ~ 0 0 0 0.1 5 normal

function rpg:monster/hurt/elemet/thunder/main_thunder

scoreboard players tag @s remove mob.target

effect @s minecraft:instant_damage 1 100
effect @s minecraft:instant_health 1 1

summon minecraft:lightning_bolt ~ ~-9 ~