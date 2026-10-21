clear @s minecraft:dye 6 1 {Tags:["材料","水之祈愿"]}
function sys:menu/chest/pray/水/pass if @s[score_stats.item_min=1]
clear @s[score_stats.item=0] clay_ball 0 0 {Tags:["材料","万化祈福碎片"]}
function sys:menu/chest/pray/水/pass_万化 if @s[score_stats.item_min=12]

function sys:menu/chest/pray/update