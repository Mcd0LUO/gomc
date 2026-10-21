clear @s bread 0 0 {display:{Lore:["§b༖§7活动道具"],Name:"§f┏⌭§b饺子§f⌭┓"},Tags:["undrop","饺子","鱼钓之季"]}
function sys:menu/chest/鱼钓之季/钓竿/专业_4 if @s[score_stats.item_min=6]
tellraw @s[score_stats.item=0] ["§r[§e鱼钓之季§r]:无效的操作，饺子不足"]