# 获取宠物主人
scoreboard players operation @a[tag=释放宠物,r=12] temp.uid.match = @s uid
execute @a[tag=释放宠物,r=20] ~ ~ ~ scoreboard players operation @s temp.uid.match -= @s uid
# 运算受到的伤害
# 整合伤害到总计
scoreboard players operation @s hurt.total += @s hurt.real
# 扣除生命值
scoreboard players operation @s health -= @s hurt.total
function rpg:monster/hurt/death/main_pet if @s[score_health=0]
# 更新血量lv/护盾提示
execute @e[c=1,dy=1,tag=attribute.hp_armor] ~ ~ ~ function rpg:monster/hurt/elemet/update_hp_armor
# 数显
scoreboard players operation @s arg1 = @s hurt.total
function sys:api/数显/main

scoreboard players reset @a[score_temp.uid.match_min=0,score_temp.uid.match=0] temp.uid.match
# reset
scoreboard players reset @s hurt.temp
scoreboard players reset @s hurt.total
scoreboard players reset @s arg1
scoreboard players reset @s arg2
scoreboard players reset @s return
scoreboard players reset @s hurt.armor
# 治疗
effect @s[tag=亡灵] minecraft:instant_damage 1 125