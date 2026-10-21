# ini
stats entity @s set AffectedItems @s stats.item
function sys:menu/chest/ini unless @s[score_stats.item_min=0]
scoreboard players set @s stats.item 0
# select
clear @s minecraft:nether_star 0 1 {Tags:["menu","淬炼.潜蛟腾跃"]}
function sys:menu/chest/淬炼/潜蛟腾跃/main if @s[score_stats.item_min=1]
# reset
stats entity @s clear AffectedItems
scoreboard players reset @s stats.item