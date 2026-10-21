# 护甲兑换接口
# ini
stats entity @s set AffectedItems @s stats.item
function sys:menu/chest/ini unless @s[score_stats.item_min=0]
scoreboard players set @s stats.item 0
# select
clear @s minecraft:leather_helmet 0 1 {Tags:["menu","露蓝头盔"]}
function sys:menu/chest/护甲/露蓝/头盔 if @s[score_stats.item_min=1]
# 灵
clear @s minecraft:leather_chestplate 0 1 {Tags:["menu","露蓝胸甲"]}
function sys:menu/chest/护甲/露蓝/胸甲 if @s[score_stats.item_min=1]
# 快速兑换
clear @s minecraft:leather_leggings 0 1 {Tags:["menu","露蓝护腿"]}
function sys:menu/chest/护甲/露蓝/护腿 if @s[score_stats.item_min=1]

clear @s minecraft:leather_boots 0 -1 {Tags:["menu","露蓝靴子"]}
function sys:menu/chest/护甲/露蓝/靴子 if @s[score_stats.item_min=1]
# reset
stats entity @s clear AffectedItems
scoreboard players reset @s stats.item