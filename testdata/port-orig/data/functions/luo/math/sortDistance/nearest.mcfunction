# 由近到远赋值
    scoreboard players add #math.nearest math.nearest 1
    scoreboard players operation @s math.nearest = #math.nearest math.nearest
    scoreboard players tag @s remove math.nearest
    execute @e[c=1,tag=math.nearest] ~ ~ ~ function luo:math/SortDistance/nearest