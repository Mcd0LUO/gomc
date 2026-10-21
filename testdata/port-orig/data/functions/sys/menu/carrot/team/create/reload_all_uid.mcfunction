# 当前无在线队长则重置当前队伍
scoreboard players reset * team.id
scoreboard players tag * remove 队长
scoreboard players tag * remove 队伍中
tellraw @s ["§r[§e系统§r]:§c检测到当前无在线队长，所有队伍id重置"]