# 钓鱼兑换接口
# ini
stats entity @s set AffectedItems @s stats.item
function sys:menu/chest/ini unless @s[score_stats.item_min=0]
scoreboard players set @s stats.item 0
# select
clear @s minecraft:fishing_rod 0 0 {Tags:["menu"]}
function sys:menu/chest/鱼钓之季/钓竿/main if @s[score_stats.item_min=1]
# 称号
clear @s minecraft:fish 3 0 {Tags:["menu"]}
function sys:menu/chest/鱼钓之季/name/main if @s[score_stats.item_min=1]
# 随机兑换
clear @s minecraft:bread 0 1 {Tags:["menu"]}
function sys:menu/chest/鱼钓之季/交换/main if @s[score_stats.item_min=1]
# 其它兑换
clear @s minecraft:book 0 0 {Tags:["menu"]}
function sys:menu/chest/鱼钓之季/其它/main if @s[score_stats.item_min=1]
# reset
stats entity @s clear AffectedItems
scoreboard players reset @s stats.item
