clear @s minecraft:dye 6 0 {Tags:["材料","水之祈愿"]}

stats entity @s clear AffectedItems
give @s[score_stats.item_min=3] leather_boots 1 0 {Unbreakable:1b,HideFlags:6,display:{Lore:["","§b⚚§7护甲[战靴]          §d۞§7命座:§4无","§e҉§f品质:§f 普通","§c⏍§7⛖▁▂▃§a§b『属性』§7▃▂▁§c⏍","§8▣ §e۞§9水盾 §a+1   §e۞§4火盾 §a+3","","§8▣ §7常见的元素护甲,可以额外抵挡§b元素伤害"],color:4088274,Name:"§f┏⌭§b露蓝华服§f⌭┓"},Tags:["ar","蓝露华服","water"]}

tellraw @s[score_stats.item=2] ["§r[§eSHOP§r]:§7购买失败,水之祈愿不足"]
tellraw @s[score_stats.item_min=3] ["§r[§eSHOP§r]:§7购买成功"]
give @s[score_stats.item_min=3] flint 1 0 {Unbreakable:1b,HideFlags:6,display:{Lore:["","§b⚚§7必备物品[永久]","","§8▣ §7手持穿戴石攻击怪物以§e重载自身元素抵抗"],Name:"§f┏⌭§6防具穿戴石§f[靴子]§f⌭┓"},Tags:["穿戴石","靴子"]}
clear @s[score_stats.item_min=3] minecraft:paper 0 64 {Tags:["材料","布鲁碎片"]}
clear @s[score_stats.item_min=3] minecraft:dye 6 3 {Tags:["材料","水之祈愿"]}
stats entity @s set AffectedItems @s stats.item