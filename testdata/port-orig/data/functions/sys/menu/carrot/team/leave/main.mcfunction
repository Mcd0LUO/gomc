# 玩家离开队伍
execute @a[tag=队长] ~ ~ ~ scoreboard players operation @s team.temp = @s team.id
scoreboard players operation @a[tag=队长] team.temp -= @s team.id
tellraw @a[score_team.temp_min=0,score_team.temp=0,tag=队长] ["§r[§e系统§r]:",{"selector":"@s"},"§7离开队伍"]
tellraw @s ["§r[§e系统§r]:§7离开了",{"selector":"@a[score_team.temp_min=0,score_team.temp=0,tag=队长]"},"队伍"]
scoreboard players tag @s remove 队伍中
scoreboard players redset @s team.id
scoreboard players reset @s menu.trigger