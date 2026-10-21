clear @s bread 0 0 {Tags:["饺子"]}
function sys:menu/chest/鱼钓之季/交换/pass if @s[score_stats.item_min=128]
tellraw @s[score_stats.item=127] ["§r[§e鱼钓之季§r]:§7购买§e失败,饺子不足"]

function sys:menu/chest/鱼钓之季/update