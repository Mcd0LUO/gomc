stats entity @s set AffectedItems @s stats.item
scoreboard players set @s stats.item 0

clear @s[score_menu.trigger_min=201,score_menu.trigger=201] paper 0 0 {Tags:["货币","末法点","1"]}
tellraw @s[score_stats.item=0,score_menu.trigger_min=201,score_menu.trigger=201] ["§r[§eBANK§r]:§7失败的存入 背包拥有§c0§7个符合条件的物品"]
function sys:menu/carrot/money/save_1 if @s[score_menu.trigger_min=201,score_menu.trigger=201,score_stats.item_min=1]

clear @s[score_menu.trigger_min=202,score_menu.trigger=202] paper 0 0 {Tags:["货币","末法点","50"]}
function sys:menu/carrot/money/save_50 if @s[score_menu.trigger_min=202,score_menu.trigger=202,score_stats.item_min=1]
tellraw @s[score_stats.item=0,score_menu.trigger_min=202,score_menu.trigger=202] ["§r[§eBANK§r]:§7失败的存入 背包拥有§c0§7个符合条件的物品"]

clear @s[score_menu.trigger_min=203,score_menu.trigger=203] paper 0 0 {Tags:["货币","末法点","100"]}
tellraw @s[score_stats.item=0,score_menu.trigger_min=203,score_menu.trigger=203] ["§r[§eBANK§r]:§7失败的存入 背包拥有§c0§7个符合条件的物品"]
function sys:menu/carrot/money/save_100 if @s[score_menu.trigger_min=203,score_menu.trigger=203,score_stats.item_min=1]

clear @s[score_menu.trigger_min=204,score_menu.trigger=204] paper 0 0 {Tags:["货币","末法点","1000"]}
tellraw @s[score_stats.item=0,score_menu.trigger_min=204,score_menu.trigger=204] ["§r[§eBANK§r]:§7失败的存入 背包拥有§c0§7个符合条件的物品"]
function sys:menu/carrot/money/save_1000 if @s[score_menu.trigger_min=204,score_menu.trigger=204,score_stats.item_min=1]

scoreboard players reset @s stats.item
stats entity @s clear AffectedItems @s