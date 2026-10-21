scoreboard players reset @s elemet.saveTotal
entitydata @s {CustomName:""}
execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s arg1 *= 2 num

execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s hurt.armor * @s 盾伤倍率

execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s armor.all.now -= @s hurt.armor

particle largeexplode ~ ~-1 ~ 0 0 0 0.25 1 normal
particle fireworksSpark ~ ~ ~ 0 0 0 0.25 10 normal

particle sweepAttack ~0.0 ~ ~-2.0 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.147153 ~ ~-1.638304 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.879385 ~ ~-0.68404 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.931852 ~ ~0.517638 -0.0 0 -0.0 0.3 1 normal
particle sweepAttack ~1.285575 ~ ~1.532089 -0.0 0 -0.0 0.3 1 normal
particle sweepAttack ~0.174311 ~ ~1.992389 -0.0 0 -0.0 0.3 1 normal
particle sweepAttack ~-1.0 ~ ~1.732051 0.0 0 -0.0 0.3 1 normal
particle sweepAttack ~-1.812616 ~ ~0.845237 0.0 0 -0.0 0.3 1 normal
particle sweepAttack ~-1.969616 ~ ~-0.347296 0.0 0 0.0 0.3 1 normal
particle sweepAttack ~-1.414214 ~ ~-1.414214 0.0 0 0.0 0.3 1 normal
particle sweepAttack ~-0.347296 ~ ~-1.969616 0.0 0 0.0 0.3 1 normal

execute @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ~ ~ ~ function rpg:monster/hurt/normal/fire/backToPlayer

tellraw @a[score_temp.uid.match_min=0,score_temp.uid.match=0] [["§7[§c§l༗ §r| §9§l☁§7]->蒸发 伤害*1.5"]]
scoreboard players tag @s remove elemet.saveWater