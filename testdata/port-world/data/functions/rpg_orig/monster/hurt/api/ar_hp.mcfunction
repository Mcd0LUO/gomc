scoreboard players operation @s return = @e[c=1,dy=-2,tag=mob.asg] health
scoreboard players operation @s return *= 100 num
scoreboard players operation @s return /= @e[c=1,dy=-2,tag=mob.asg] health.max
function rpg:monster/hurt/api/health_lv/main
scoreboard players operation @s return = @e[c=1,dy=-2,tag=mob.asg] armor.all.now
scoreboard players operation @s return *= 100 num
scoreboard players operation @s return /= @e[c=1,dy=-2,tag=mob.asg] armor.all.max
function rpg:monster/hurt/api/armor_lv/main

scoreboard players reset @s return