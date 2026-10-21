scoreboard players tag @s add 吸附光环
tellraw @s ["§r[§e鱼钓之季§r]:§7购买§e成功,钓鱼快乐啊"]
clear @s bread 0 256 {Tags:["undrop","饺子","鱼钓之季"]}
clear @s pumpkin_pie 0 8 {Tags:["undrop","小龙糕","鱼钓之季"]}
clear @s gold_ingot 0 3 {Tags:["undrop","龙年大吉","鱼钓之季"]}
function sys:api/粒子/灵子崩坏/+y
tellraw @a ["§r[§e鱼钓之季§r]",{"selector":"@s"},":§7成功购买§e吸附§7光环,挂机不愁~,钓鱼快乐啊~"]
scoreboard players tag @s remove 饺子充足
scoreboard players tag @s remove 小龙鱼充足
scoreboard players tag @s remove 小龙糕充足
scoreboard players tag @s remove 龙年大吉充足
scoreboard players tag @s add 兑换成功
