scoreboard players operation @e[tag=pet] pet.temp = @s uid
execute @e[tag=pet] ~ ~ ~ scoreboard players operation @s pet.temp -= @s pet.uid
teleport @e[tag=pet,score_pet.temp_min=0,score_pet.temp=0] ~ ~ ~0.5
execute @s[tag=释放宠物] ~ ~ ~ function sys:menu/carrot/pet/sum_list/抛异 unless @e[tag=pet,score_pet.temp_min=0,score_pet.temp=0]
scoreboard players reset @e[tag=pet] pet.temp

tellraw @s[tag=释放宠物] ["§e✦ §7战兽召回成功"]
tellraw @s[tag=!释放宠物] ["§c✘ §eERROR:§7当前无活动的契约战兽"]