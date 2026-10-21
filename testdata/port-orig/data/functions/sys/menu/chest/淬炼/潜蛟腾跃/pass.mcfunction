function sys:math/random/1-100
tellraw @s[score_return_min=1,score_return=40] ["§r[§e洛必达§r]:§7匠神的恩赐没有降下,§7淬炼提升§c失败,物品销毁*1"]
tellraw @s[score_return_min=41,score_return=100] ["§r[§e洛必达§r]:§7感受到§e匠神§7的恩赐下,§r淬炼提升§e成功"]
clear @s nether_star 0 1 {Tags:["潜蛟腾跃","栏位:1","淬炼:0"]}
give @s[score_return_min=41,score_return=100] minecraft:nether_star 1 0 {display:{Lore:["§b⚚§7星纹          §d۞§7淬炼:§41§d","§c⏍§7⛖▁▂▃§a§b『属性』§7▃▂▁§c⏍","§8▣ §eMP回复҉ : 8/3s","§8▣ §cHP✙回复 : 3/3s","","§8▣ §c§7许愿池中水龙灵精魂所化,给人安宁的力量","§8▣ §c§7可在§e[星升之铺]§7提升淬炼等级","§8▣ §c§7需佩戴在物品栏第一行第一格[从上往下数]"],Name:"§f┏⌭§b潜蛟腾跃§f⌭┓"},Tags:["星纹","undrop","潜蛟腾跃","栏位:1","淬炼:1"]}