replaceitem block 123 31 -154 slot.container.0 minecraft:paper 1 0 {display:{Name:"§e1 末法点§c*1",Lore:["§7售价:§a布鲁碎片§c*1","","§7单击兑换一次","","§7快速兑换请点击右下角拉杆"]},Tags:["menu","末法点","1"]}

replaceitem block 123 31 -154 slot.container.1 minecraft:paper 1 0 {display:{Name:"§e50 末法点§c*1",Lore:["§7售价:§a孤落遗章§c*1","","§7单击兑换一次","","§7快速兑换请点击右下角拉杆","","§7兑换增加§b全服§e孤落法典§7进度 §a1","","§7全服进度达2000即可全服发放"]},Tags:["menu","末法点","50"]}

#replaceitem block 123 31 -154 slot.container.9 minecraft:dye 1 4 {display:{Lore:["§7售价:§a虚拟末法点§c*850","","§7单击兑换一次","","§7快速兑换:关闭","","§b༖§7功能屋穿戴/脱下护甲"],Name:"§f┏⌭§b灵§f⌭┓"},Tags:["menu","材料","灵","dye"]}


execute @s[tag=!快速兑换] ~ ~ ~ replaceitem block 124 31 -154 slot.container.26 minecraft:lever 1 0 {display:{Name:"§d快速兑换§c[关闭]",Lore:["§e灯亮§7开启快速兑换","","§c灯灭§7关闭快速兑换","","§7快速兑换:将背包中§b所有§a匹配资源§7购买当前§e选中§7的物品","","§7每次购买后需重新开启"]},Tags:["menu","快速兑换","off"]}

execute @s[tag=快速兑换] ~ ~ ~ replaceitem block 124 31 -154 slot.container.26 minecraft:redstone_torch 1 0 {display:{Name:"§d快速兑换§e[开启]",Lore:["§e灯亮§7开启快速兑换","","§c灯灭§7关闭快速兑换","","§7快速兑换:将背包中§b所有§a匹配资源§7购买当前§e选中§7的物品","","§7每次购买后需重新开启"]},Tags:["menu","快速兑换","on"]}