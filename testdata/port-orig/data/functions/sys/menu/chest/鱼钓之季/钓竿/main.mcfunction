# select
clear @s minecraft:fishing_rod 0 1 {Tags:["menu","普通钓竿"]}
function sys:menu/chest/鱼钓之季/钓竿/普通 if @s[score_stats.item_min=1]
clear @s minecraft:fishing_rod 0 1 {Tags:["menu","专业钓竿"]}
function sys:menu/chest/鱼钓之季/钓竿/专业 if @s[score_stats.item_min=1]
clear @s minecraft:fishing_rod 0 1 {Tags:["menu","大师钓竿"]}
function sys:menu/chest/鱼钓之季/钓竿/大师 if @s[score_stats.item_min=1]
clear @s minecraft:fishing_rod 0 1 {Tags:["menu","传奇钓竿"]}
function sys:menu/chest/鱼钓之季/钓竿/传奇 if @s[score_stats.item_min=1]

function sys:menu/chest/鱼钓之季/update