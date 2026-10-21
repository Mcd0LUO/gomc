clear @s fish 2 0 {display:{Lore:["§b༖§7活动道具"],Name:"§f┏⌭§e小龙鱼§f⌭┓"},Tags:["undrop","小龙鱼","鱼钓之季"]}
function sys:menu/chest/鱼钓之季/钓竿/专业_3 if @s[score_stats.item_min=1]
tellraw @s[score_stats.item=0] ["§r[§e鱼钓之季§r]:无效的操作，小龙鱼不足"]