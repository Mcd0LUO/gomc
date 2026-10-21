# 创建队伍

# 重置所有队伍
function sys:menu/carrot/team/create/reload_all_uid unless @a[tag=队长]

scoreboard players add #team.id# team.id 1
scoreboard players operation @s team.id = #team.id# team.id
tellraw @s ["§r[§e系统§r]:§7创建成功，队伍ID.",{"score":{"objective":"team.id","name":"@s"},"color":"red"}]
scoreboard players tag @s add 队长
scoreboard players tag @s add 队伍中
scoreboard players set @s menu.trigger 0