execute @s[score_temp_min=1,score_temp=1] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l1 "}
execute @s[score_temp_min=2,score_temp=2] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l2 "}
execute @s[score_temp_min=3,score_temp=3] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l3 "}
execute @s[score_temp_min=4,score_temp=4] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l4 "}
execute @s[score_temp_min=5,score_temp=5] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l5 "}
execute @s[score_temp_min=6,score_temp=6] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l6 "}
execute @s[score_temp_min=7,score_temp=7] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l7 "}
execute @s[score_temp_min=8,score_temp=8] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l8 "}
execute @s[score_temp_min=9,score_temp=9] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l9 "}
execute @s[score_temp_min=0,score_temp=0] ~ ~ ~ entitydata @e[c=1,tag=数显.3] {CustomName:"§c§l0 "}

function sys:api/数显/scale_千 if @s[score_arg1_min=1000]
scoreboard players tag @e[c=1,tag=数显.3] remove 数显.3