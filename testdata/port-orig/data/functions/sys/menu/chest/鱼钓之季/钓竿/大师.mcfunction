clear @s bread 0 0 {Tags:["undrop","饺子","鱼钓之季"]}
scoreboard players tag @s[score_stats.item_min=32] add 饺子充足
clear @s fish 2 0 {Tags:["undrop","小龙鱼","鱼钓之季"]}
scoreboard players tag @s[score_stats.item_min=10] add 小龙鱼充足
clear @s pumpkin_pie 0 0 {Tags:["undrop","小龙糕","鱼钓之季"]}
scoreboard players tag @s[score_stats.item_min=8] add 小龙糕充足

execute @s[tag=饺子充足] ~ ~ ~ execute @s[tag=小龙鱼充足] ~ ~ ~ execute @s[tag=小龙糕充足] ~ ~ ~ function sys:menu/chest/鱼钓之季/钓竿/大师_2
tellraw @s[tag=!兑换成功] ["§r[§e鱼钓之季§r]:§7兑换失败"]
scoreboard players tag @s remove 兑换成功
scoreboard players tag @s remove 饺子充足
scoreboard players tag @s remove 小龙鱼充足
scoreboard players tag @s remove 小龙糕充足