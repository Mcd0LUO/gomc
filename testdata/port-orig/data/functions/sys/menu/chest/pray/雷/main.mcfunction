clear @s minecraft:dye 11 1 {Tags:["材料","雷之祈愿"]}
function sys:menu/chest/pray/雷/pass if @s[score_stats.item_min=1]
tellraw @s[score_stats.item=0] ["§r[§e歌莉娅§r]:§7你还没有§e雷之祈愿§7哦，请前往副本战胜§e雷§7属性怪物吧"]


function sys:menu/chest/pray/update