tellraw @s ["§r[§e鱼钓之季§r]:§7购买§e成功,钓鱼快乐啊"]
give @s lime_shulker_box 1 0 {display:{Lore:["§e功能屋使用"],Name:"§f┏⌭§b额外背包§f⌭┓"}}

clear @s bread 0 256 {Tags:["undrop","饺子","鱼钓之季"]}
clear @s fish 2 32 {Tags:["undrop","小龙鱼","鱼钓之季"]}
clear @s pumpkin_pie 0 32 {Tags:["undrop","小龙糕","鱼钓之季"]}
clear @s gold_ingot 0 1 {Tags:["undrop","龙年大吉","鱼钓之季"]}
function sys:api/粒子/灵子崩坏/+y
scoreboard players tag @s add 兑换成功
