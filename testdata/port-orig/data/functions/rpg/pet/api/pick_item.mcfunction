scoreboard players operation @a[tag=释放宠物] pet.temp = @s uid
execute @a[tag=释放宠物] ~ ~ ~ scoreboard players operation @s pet.temp -= @s uid
tp @e[r=2,type=item] @a[score_pet.temp=0,score_pet.temp=0]
scoreboard players reset @a[tag=释放宠物] pet.temp