# match
scoreboard players operation @e[tag=skill.岩狱之路] skill.temp = @s uid
scoreboard players operation @a[r=20] skill.temp2 = @s uid
execute @e[tag=skill.岩狱之路] ~ ~ ~ scoreboard players operation @s skill.temp -= @s uid
execute @a[r=20] ~ ~ ~ scoreboard players operation @s skill.temp2 -= @s uid
scoreboard players set @a[r=20,score_skill.temp2_min=0,score_skill.temp2=0] temp.uid.match 0
# 怪物
scoreboard players operation @e[r=2,tag=mob.asg] hurt.fire = @s damage.fire
execute @e[r=2,tag=mob.asg] ~ ~ ~ function rpg:monster/hurt/elemet/fire/main_fire
# reset
kill @e[tag=skill.岩狱之路,score_skill.temp_min=0,score_skill.temp=0]
scoreboard players reset @a[r=20,score_skill.temp2_min=0,score_skill.temp2=0] skill.id
scoreboard players reset @a[r=20,score_skill.temp2_min=0,score_skill.temp2=0] skill.time.fire
scoreboard players reset @a[r=20,score_skill.temp2_min=0,score_skill.temp2=0] skill.temp2