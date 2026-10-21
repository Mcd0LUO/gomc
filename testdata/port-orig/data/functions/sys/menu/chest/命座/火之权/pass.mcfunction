scoreboard players add @s damage.armor 5
scoreboard players add @s more.fire 15
scoreboard players remove @s money 1500
tellraw @s ["§r[§e洛必达]:§7命座提升成功"]
clear @s carrot_on_a_stick 0 1 {Unbreakable:1b,HideFlags:4,display:{Lore:["","§b⚚§7触发杆          §d۞§7命座:§40§d","§c⏍§7⛖▁▂▃§a§b『属性』§7▃▂▁§c⏍","§8▣ §e⚔ §a3 + §6元素等级","§8▣ §3§l⚡MP: §a5/一次","","§8▣ §7右键释放§4火元素§7攻击 每次积攒§e怒气值 §a5","§8▣ §7可在§e光逝之门§7提升命座","§8▣ §7在短时间内按下§eShift+F键§7可进行振刀完全抵挡伤害"],Name:"§f┏⌭§4火之权§f⌭┓"},Tags:["wp","hzq","fire"]}

function sys:misc/give/weapon/火之权1