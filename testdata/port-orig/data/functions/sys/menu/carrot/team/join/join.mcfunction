execute @a[tag=队长] ~ ~ ~ scoreboard players operation @s team.temp = @s team.key
scoreboard players operation @a[tag=队长] team.temp -= @s team.join
scoreboard players operation @s team.id = @a[score_team.temp_min=0,score_team.temp=0] team.id
scoreboard players tag @s add 队伍中
tellraw @s ["§r[§e系统§r]:§7已加入",{"selector":"@a[score_team.temp_min=0,score_team.temp=0]","color": "red"},"§7的队伍"]
tellraw @a[score_team.temp_min=0,score_team.temp=0] ["§r[§e系统§r]:",{"selector":"@s","color":"red"},"§7加入了你的队伍"]
scoreboard players reset @s team.join
scoreboard players reset @a[score_team.temp_min=1] team.temp
scoreboard players reset @a team.key
scoreboard players reset @s menu.trigger