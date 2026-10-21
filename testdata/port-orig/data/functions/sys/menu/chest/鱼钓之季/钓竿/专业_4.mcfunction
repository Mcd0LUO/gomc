function sys:misc/give/活动/鱼钓之季/专业钓竿
tellraw @s ["§r[§e鱼钓之季§r]:§7成功购买§e专业§7钓竿,钓鱼快乐啊~"]
clear @s bread 0 6 {display:{Lore:["§b༖§7活动道具"],Name:"§f┏⌭§b饺子§f⌭┓"},Tags:["undrop","饺子","鱼钓之季"]}
clear @s fish 2 1 {display:{Lore:["§b༖§7活动道具"],Name:"§f┏⌭§e小龙鱼§f⌭┓"},Tags:["undrop","小龙鱼","鱼钓之季"]}

function sys:api/粒子/灵子崩坏/+y