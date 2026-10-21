# 选择副本进入

clear @s minecraft:knowledge_book 0 -1 {Tags:["menu","孤落之森"]}
function sys:menu/chest/tp_c/孤落之森 if @s[score_stats.item_min=1]

clear @s minecraft:knowledge_book 0 -1 {Tags:["menu","凝冰之原"]}
function sys:menu/chest/tp_c/凝冰之原 if @s[score_stats.item_min=1]

# 更新
function sys:menu/chest/tp_c/update
function sys:menu/carrot/pet/sum_list/召回 if @s[tag=释放宠物]