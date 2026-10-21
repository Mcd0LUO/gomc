# armor
scoreboard players operation @s armor.all.now = @e[c=1,tag=mob.asg,dy=-1] armor.all.now
scoreboard players operation @s armor.all.max = @e[c=1,tag=mob.asg,dy=-1] armor.all.max
scoreboard players operation @s return = @s armor.all.now
scoreboard players operation @s return *= 100 num
scoreboard players operation @s return /= @s armor.all.max
function rpg:monster/hurt/api/armor_lv/main if @s[score_return_min=1]
scoreboard teams join 0 @s[score_return=0]
# health
scoreboard players operation @s health = @e[c=1,tag=mob.asg,dy=-1] health
scoreboard players operation @s health.max = @e[c=1,tag=mob.asg,dy=-1] health.max
scoreboard players operation @s return = @s health
scoreboard players operation @s return *= 100 num
scoreboard players operation @s return /= @s health.max
function rpg:monster/hurt/api/health_lv/main
scoreboard players reset @s health
scoreboard players reset @s armor.all.now