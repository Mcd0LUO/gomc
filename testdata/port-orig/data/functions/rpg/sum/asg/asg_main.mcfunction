# 赋值
function rpg:sum/asg/1c_main if @s[tag=1c]

function rpg:sum/asg/2c_main if @s[tag=2c]

function rpg:sum/asg/boss_main if @s[tag=boss]
# 特殊标记
scoreboard players operation @s health.max = @s health
scoreboard players tag @s remove new
scoreboard players tag @s add mob.asg
scoreboard teams join 100 @e[dy=2,tag=attribute.hp_armor]