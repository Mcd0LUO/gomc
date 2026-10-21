function sys:misc/give/活动/鱼钓之季/大师钓竿

scoreboard players tag @s add 兑换成功
clear @s bread 0 32 {Tags:["undrop","饺子","鱼钓之季"]}
clear @s fish 2 6 {Tags:["undrop","小龙鱼","鱼钓之季"]}
clear @s pumpkin_pie 0 6 {Tags:["undrop","小龙糕","鱼钓之季"]}

tellraw @s ["§r[§e鱼钓之季§r]:§7成功购买§e大师§7钓竿,钓鱼快乐啊~"]
