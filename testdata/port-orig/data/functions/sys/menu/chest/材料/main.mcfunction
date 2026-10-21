# 材料兑换接口

# ini
stats entity @s set AffectedItems @s stats.item
function sys:menu/chest/ini unless @s[score_stats.item_min=0]
scoreboard players set @s stats.item 0
# select
clear @s minecraft:paper 0 0 {Tags:["menu","末法点"]}
function sys:menu/chest/材料/末法点/main if @s[score_stats.item_min=1]
# 快速兑换
clear @s minecraft:lever 0 -1 {Tags:["menu","快速兑换","off"]}
function sys:menu/chest/材料/快速兑换/on if @s[score_stats.item_min=1]
clear @s minecraft:redstone_torch 0 -1 {Tags:["menu","快速兑换","on"]}
function sys:menu/chest/材料/快速兑换/off if @s[score_stats.item_min=1]
# reset
stats entity @s clear AffectedItems
scoreboard players reset @s stats.item