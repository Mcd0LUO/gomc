scoreboard players operation @s bar.cold -= #1 ggm.c
scoreboard players operation @s t2.ggm = @s bar.cold
scoreboard players operation @s t2.ggm > #0 ggm.c
scoreboard players operation @s bar.cold = @s t2.ggm
scoreboard players operation @s t2.ggm = @s bar.cold
scoreboard players operation @s t2.ggm -= #50 ggm.c
function rpg:attribute/cold/main$b1 if @s[score_t2.ggm=0]
