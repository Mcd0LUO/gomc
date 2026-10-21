# 副本传送箱子菜单接口

# ini
stats entity @s set AffectedItems @s stats.item
function sys:menu/chest/ini unless @s[score_stats.item_min=0]
scoreboard players set @s stats.item 0
scoreboard players reset @s data.walk
# select
clear @s minecraft:knowledge_book 0 0 {Tags:["menu"]}
function sys:menu/chest/tp_c/select if @s[score_stats.item_min=1]

# reset
stats entity @s clear AffectedItems
scoreboard players reset @s stats.item