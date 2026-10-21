# __初始化第一次登录的玩家
tp @s 84 9 -163
spawnpoint @s ~ ~ ~
#tellraw @s ["§r[§e系统§r]:§7欢迎进入服务器，请选择你的职业"]
function sys:basic/player_ini/uid/main
# 初始化属性
scoreboard players set @s money 0
scoreboard players set @s health 20
scoreboard players set @s health.max 20
scoreboard players set @s armor.all 0
scoreboard players set @s mp.now 0
scoreboard players set @s mp.max 100
scoreboard players set @s damage.fire 0
scoreboard players set @s damage.water 0
scoreboard players set @s armor.player.all 0
scoreboard players set @s elemet.fire_lv 0
scoreboard players set @s elemet.water_lv 0
scoreboard players set @s elemet.ice_lv 0
scoreboard players set @s elemet.thund_lv 0
scoreboard players set @s dq 0
scoreboard players set @s plot.time 0