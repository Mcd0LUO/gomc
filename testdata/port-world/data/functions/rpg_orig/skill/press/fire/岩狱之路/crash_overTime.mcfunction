scoreboard players operation @e[tag=skill.岩狱之路] skill.temp = @s uid
execute @e[tag=skill.岩狱之路] ~ ~ ~ scoreboard players operation @s skill.temp -= @s uid
execute @a[r=20] ~ ~ ~ scoreboard players operation @s skill.temp2 -= @s uid

kill @e[tag=skill.岩狱之路,score_skill.temp_min=0,score_skill.temp=0]
scoreboard players reset @s skill.id
scoreboard players reset @s skill.time.fire
scoreboard players reset @s skill.temp2