scoreboard players operation @s arg1 = @s money 
scoreboard players remove @s arg1 1500
scoreboard players tag @s[score_elemet.water_lv_min=10] add 允许升级 {SelectedItem:{tag:{Tags:["szq"]}}}
scoreboard players tag @s remove 允许升级 {SelectedItem:{tag:{Tags:["命座"]}}}
function sys:menu/chest/命座/水之权/pass if @s[score_arg1_min=0,tag=允许升级]
function sys:menu/chest/命座/水之权/fail if @s[tag=!允许升级]

function sys:menu/chest/命座/update
scoreboard players reset @s arg1
scoreboard players tag @s remove 允许升级