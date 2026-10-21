function sys:skill/press/fire/岩狱之路/ini unless @s[score_skill.time.fire_min=0]

scoreboard players add @s skill.time.fire 1

scoreboard players operation @s skill.temp = @s skill.time.fire
scoreboard players operation @s skill.temp %= 2 num

# execute @s[score_skill.temp_min=0,score_skill.temp=0] ~ ~ ~ 
execute @e[c=1,tag=skill.岩狱之路.1] ~ ~ ~ function sys:skill/press/fire/岩狱之路/step_1

execute @e[type=armor_stand,tag=skill.岩狱之路.2] ~ ~ ~ function sys:skill/press/fire/岩狱之路/hurt if @e[tag=mob.asg,r=2]

scoreboard players reset @s skill.temp

function sys:skill/press/fire/岩狱之路/crash_overTime if @s[score_skill.time.fire_min=20]