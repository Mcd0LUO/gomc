clear @s minecraft:paper 0 -1 {Tags:["menu","末法点","1"]}
function sys:menu/chest/材料/末法点/1 if @s[score_stats.item_min=1,tag=!快速兑换]
function sys:menu/chest/材料/末法点/1_fast if @s[score_stats.item_min=1,tag=快速兑换]
clear @s minecraft:paper 0 -1 {Tags:["menu","末法点","50"]}
function sys:menu/chest/材料/末法点/50 if @s[score_stats.item_min=1,tag=!快速兑换]
function sys:menu/chest/材料/末法点/50_fast if @s[score_stats.item_min=1,tag=快速兑换]

function sys:menu/chest/材料/update