clear @s minecraft:book 0 -1 {Tags:["材料","孤落遗章"]}
function sys:menu/chest/材料/末法点/50_fast_pass if @s[score_stats.item_min=1]
tellraw @s[score_stats.item=0] ["§r[§eSHOP§r]:§7购买失败"]
scoreboard players reset @s data.walk