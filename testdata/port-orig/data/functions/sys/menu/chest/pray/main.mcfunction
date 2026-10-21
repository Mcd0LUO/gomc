# ini
stats entity @s set AffectedItems @s stats.item
function sys:menu/chest/ini unless @s[score_stats.item_min=0]
scoreboard players set @s stats.item 0
# select
clear @s minecraft:dye 1 -1 {Tags:["menu","火元素祈福"]}
function sys:menu/chest/pray/火/main if @s[score_stats.item_min=1]

clear @s minecraft:dye 6 -1 {Tags:["menu","水元素祈福"]}
function sys:menu/chest/pray/水/main if @s[score_stats.item_min=1]

clear @s minecraft:dye 11 -1 {Tags:["menu","雷元素祈福"]}
function sys:menu/chest/pray/雷/main if @s[score_stats.item_min=1]

clear @s minecraft:dye 12 -1 {Tags:["menu","冰元素祈福"]}
function sys:menu/chest/pray/冰/main if @s[score_stats.item_min=1]

# reset
stats entity @s clear AffectedItems
scoreboard players reset @s stats.item