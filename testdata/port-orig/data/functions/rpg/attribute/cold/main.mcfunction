scoreboard players remove @s bar.cold 1
scoreboard players operation @s bar.cold > 0 num
function sys:attribute/cold/debuff_main if @s[score_bar.cold=50]