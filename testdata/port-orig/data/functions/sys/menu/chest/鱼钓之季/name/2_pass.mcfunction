scoreboard players tag @s add 兑换成功
clear @s bread 0 64 {Tags:["undrop","饺子","鱼钓之季"]}
clear @s fish 2 32 {Tags:["undrop","小龙鱼","鱼钓之季"]}
clear @s pumpkin_pie 0 22 {Tags:["undrop","小龙糕","鱼钓之季"]}
clear @s gold_ingot 0 6 {Tags:["undrop","龙年大吉","鱼钓之季"]}

tellraw @a ["§r[§e鱼钓之季§r]",{"selector":"@s"},":§7成功兑换§e称号:承包鱼塘II§7已经超越了大多数人辣,钓鱼快乐啊~"]
tellraw @s ["§r[§e鱼钓之季§r]",{"selector":"@s"},":§7称号已发放至称号仓库,钓鱼快乐啊~"]
scoreboard players tag @s remove 饺子充足
scoreboard players tag @s remove 小龙鱼充足
scoreboard players tag @s remove 小龙糕充足
scoreboard players tag @s remove 龙年大吉充足
tp @s ~ ~ ~ ~ -90

# 给予称号
scoreboard players set @s name.承包鱼塘 2

function sys:api/粒子/灵子崩坏/+y