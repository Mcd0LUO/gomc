# 技能遇怪伤害赋值
scoreboard players operation @e[c=1,tag=mob.asg] uid = @s uid
scoreboard players operation @e[c=1,tag=mob.asg] hurt.ice = @s damage.ice
scoreboard players operation @e[c=1,tag=mob.asg] hurt.real = @s damage.real
execute @e[c=1,tag=mob.asg] ~ ~ ~ function rpg:monster/hurt/elemet/ice/main_ice
kill @s