# ini
stats entity @s set AffectedItems @s stats.item
function sys:menu/chest/ini unless @s[score_stats.item_min=0]
scoreboard players set @s stats.item 0
# select
clear @s minecraft:carrot_on_a_stick 0 1 {Tags:["menu","命座.火之权"]}
function sys:menu/chest/命座/火之权/main if @s[score_stats.item_min=1]
clear @s minecraft:carrot_on_a_stick 0 1 {Tags:["menu","命座.水之权"]}
function sys:menu/chest/命座/水之权/main if @s[score_stats.item_min=1]
clear @s minecraft:carrot_on_a_stick 0 1 {Tags:["menu","命座.冰之权"]}
function sys:menu/chest/命座/冰之权/main if @s[score_stats.item_min=1]
clear @s minecraft:carrot_on_a_stick 0 1 {Tags:["menu","命座.雷之权"]}
function sys:menu/chest/命座/雷之权/main if @s[score_stats.item_min=1]
# reset
stats entity @s clear AffectedItems
scoreboard players reset @s stats.item