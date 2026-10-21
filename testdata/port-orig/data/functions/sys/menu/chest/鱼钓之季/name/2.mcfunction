clear @s bread 0 0 {Tags:["undrop","饺子","鱼钓之季"]}
scoreboard players tag @s[score_stats.item_min=64] add 饺子充足
clear @s fish 2 0 {Tags:["undrop","小龙鱼","鱼钓之季"]}
scoreboard players tag @s[score_stats.item_min=32] add 小龙鱼充足
clear @s pumpkin_pie 0 0 {Tags:["undrop","小龙糕","鱼钓之季"]}
scoreboard players tag @s[score_stats.item_min=22] add 小龙糕充足
clear @s gold_ingot 0 0 {Tags:["undrop","龙年大吉","鱼钓之季"]}
scoreboard players tag @s[score_stats.item_min=6] add 龙年大吉充足

scoreboard players tag @s[score_name.承包鱼塘_min=2] remove 饺子充足
execute @s[tag=饺子充足] ~ ~ ~ execute @s[tag=小龙鱼充足] ~ ~ ~ execute @s[tag=小龙糕充足] ~ ~ ~ execute @s[tag=龙年大吉充足] ~ ~ ~ function sys:menu/chest/鱼钓之季/name/2_pass
tellraw @s[tag=!兑换成功] ["§r[§e鱼钓之季§r]:§7兑换失败 条件不符合或已拥有该等级称号"]

playsound entity.experience_orb.pickup master @s ~0 ~0 ~0 1 1.5 0
particle fireworksSpark ~ ~ ~ 0 0 0 0.25 10 normal
particle totem ~ ~ ~ 0 0 0 0.25 10 normal


scoreboard players tag @s remove 饺子充足
scoreboard players tag @s remove 小龙鱼充足
scoreboard players tag @s remove 小龙糕充足
scoreboard players tag @s remove 龙年大吉充足
scoreboard players tag @s remove 兑换成功