scoreboard players operation @s t3.ggm = @s bar.cold
scoreboard players operation @s t3.ggm -= #20 ggm.c
function rpg:attribute/cold/debuff_main$b4 unless @s[score_t3.ggm=0]
function rpg:attribute/cold/debuff_main$b3 if @s[score_t3.ggm=0]
