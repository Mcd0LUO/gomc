# __分配uid [基础算法,注释免了吧]

scoreboard players operation @s uid = #uid# uid

scoreboard players add #uid# uid 1

tellraw @a ["§r[§e系统§r]:§7新玩家",{"selector":"@s","color": "red"},"§7注册了服务器ID",{"score":{"objective":"uid","name":"@s"},"color":"red"}]

scoreboard players tag @s add login

scoreboard players set @s @s plot.id 1