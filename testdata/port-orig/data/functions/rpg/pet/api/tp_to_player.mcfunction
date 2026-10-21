scoreboard players operation @e[tag=宠物释放] pet.temp = @s uid
execute @e[tag=pet] ~ ~ ~ scoreboard players operation @s pet.temp -= @s pet.uid
tp @e[tag=pet,score_pet.temp_min=0,score_pet.temp=0,rm=20] @s
scoreboard players reset @e[tag=宠物释放] temp