# 管理队伍
scoreboard players operation @a[tag=队伍中] team.temp = @s team.id
execute @a[tag=队伍中] ~ ~ ~ scoreboard players operation @s team.temp -= @s team.id

tellraw @s [{"text":"┍───────────────管理队伍─────────────┑\n│"},{"text":"┅§c[解散队伍]§r┅","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 3007"}},{"text":" ┅§e[踢出玩家]§r┅","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 3008"}},"\n§7玩家列表:",{"selector":"@s","color":"green"}]

scoreboard players reset @a[tag=队伍中,team.temp]