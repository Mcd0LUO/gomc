# 获取1-20随机密钥 并删除该密钥获得方式
scoreboard players operation @s name.id = @r[type=armor_stand,tag=database.name,score_name.id_min=1] name.id
scoreboard players operation @e[type=armor_stand,tag=database.name] arg1 = @s name.id
execute @e[type=armor_stand,tag=database.name] ~ ~ ~ scoreboard players operation @s arg1 -= @s name.id
scoreboard players reset @e[type=armor_stand,tag=database.name,score_arg1_min=0,score_arg1=0] name.id
scoreboard players reset @e[type=armor_stand,tag=database.name] arg1