scoreboard players operation @a[tag=队伍中] team.temp = @s team.id
execute @a[tag=队伍中] ~ ~ ~ scoreboard players operation @s team.temp -= @s team.id
tellraw @a[score_team.temp_min=0,score_team.temp=0,tag=队伍中] ["§r[§e系统§r]:§7您所在队伍§c已解散"]
scoreboard players reset @a[score_team.temp_min=0,score_team.temp=0,tag=队伍中] team.id
scoreboard players tag @a[score_team.temp_min=0,score_team.temp=0,tag=队伍中] remove 队伍中
scoreboard players reset @a[score_team.temp_min=-999999] team.temp
scoreboard players reset @s menu.trigger