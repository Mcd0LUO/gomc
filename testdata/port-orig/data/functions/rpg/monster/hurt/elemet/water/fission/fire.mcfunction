scoreboard players tag @s remove elemet.saveFire
scoreboard players reset @s elemet.saveTotal
entitydata @s {CustomName:""}
execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s arg1 *= 2 num

#execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s hurt.armor *= @s 盾伤倍率

execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s armor.all.now -= @s hurt.armor

particle crit ~ ~ ~ 0 0 0 0.25 10 normal
particle angryVillager ~ ~-1 ~ 0 0 0 0.25 10 normal
particle lava ~ ~-1 ~ 0 -1 0 0.25 10 normal
particle iconcrack ~-1 ~1.5 ~ 0 0 0 0.25 20 normal @a 22
particle smoke ~ ~-1 ~ 0 0 0 0.25 10 normal
particle reddust ~0.0 ~ ~-2.0 -0.0 -1 0.0 0.3 10 normal
particle reddust ~1.0 ~ ~-1.732051 -0.0 -1 0.0 0.3 10 normal
particle reddust ~1.732051 ~ ~-1.0 -0.0 -1 0.0 0.3 10 normal
particle reddust ~2.0 ~ ~0.0 -0.0 -1 -0.0 0.3 10 normal     
particle reddust ~1.732051 ~ ~1.0 -0.0 -1 -0.0 0.3 10 normal
particle reddust ~1.0 ~ ~1.732051 -0.0 -1 -0.0 0.3 10 normal
particle reddust ~-0.0 ~ ~2.0 0.0 -1 -0.0 0.3 10 normal     
particle reddust ~-1.0 ~ ~1.732051 0.0 -1 -0.0 0.3 10 normal
particle reddust ~-1.732051 ~ ~1.0 0.0 -1 -0.0 0.3 10 normal
particle reddust ~-2.0 ~ ~-0.0 0.0 -1 0.0 0.3 1 normal
particle reddust ~-1.732051 ~ ~-1.0 0.0 -1 0.0 0.3 10 normal
particle reddust ~-1.0 ~ ~-1.732051 0.0 -1 0.0 0.3 10 normal

tellraw @a[score_temp.uid.match_min=0,score_temp.uid.match=0] [["§7[§9§l☁§r | §c§l§c§l༗§7]->§7蒸发  伤害*2"]]