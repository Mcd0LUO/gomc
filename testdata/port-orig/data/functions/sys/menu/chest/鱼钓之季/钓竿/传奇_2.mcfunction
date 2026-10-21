function sys:misc/give/活动/鱼钓之季/传奇钓竿

scoreboard players tag @s add 兑换成功
clear @s bread 0 52 {Tags:["undrop","饺子","鱼钓之季"]}
clear @s fish 2 26 {Tags:["undrop","小龙鱼","鱼钓之季"]}
clear @s pumpkin_pie 0 16 {Tags:["undrop","小龙糕","鱼钓之季"]}
clear @s gold_ingot 0 3 {Tags:["undrop","龙年大吉","鱼钓之季"]}
function sys:api/粒子/灵子崩坏/+y
tellraw @a ["§r[§e鱼钓之季§r]",{"selector":"@s"},":§7成功购买§e传奇§7钓竿,离承包鱼塘不远辣,钓鱼快乐啊~"]
scoreboard players tag @s remove 饺子充足
scoreboard players tag @s remove 小龙鱼充足
scoreboard players tag @s remove 小龙糕充足
scoreboard players tag @s remove 龙年大吉充足