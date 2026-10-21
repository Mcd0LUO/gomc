# 怪物受冰元素攻击主接口
# 受击反馈

# 获取攻击玩家
scoreboard players operation @a[r=12] temp.uid.match = @s uid
execute @a[r=12] ~ ~ ~ scoreboard players operation @s temp.uid.match -= @s uid
# 运算受到的伤害
execute @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ~ ~ ~ function rpg:monster/hurt/normal/ice/backToPlayer
# 整合伤害到总计
scoreboard players operation @s hurt.total += @s hurt.ice
# 护盾阶段判断[结果——>hurt.temp]
scoreboard players operation @s armor.all.now -= @s hurt.armor
scoreboard players operation @s arg1 = @s armor.all.now
scoreboard players operation @s arg1 *= 100 num
scoreboard players operation @s arg1 /= @s armor.all.max
scoreboard players set @s[score_arg1_min=80] hurt.temp 80
scoreboard players set @s[score_arg1_min=40,score_arg1=79] hurt.temp 50
scoreboard players set @s[score_arg1_min=1,score_arg1=39] hurt.temp 20
scoreboard players set @s[score_arg1=0] hurt.temp 0
# 额外削弱本系伤害【护盾+10%】
scoreboard players add @s[tag=ice] hurt.temp 10
scoreboard players remove @s[tag=fire] hurt.temp 10
# 最终减伤 arg1为结果
scoreboard players set @s arg1 0
scoreboard players operation @s arg1 = @s hurt.total
scoreboard players set @s arg2 100
scoreboard players operation @s arg2 -= @s hurt.temp
scoreboard players operation @s arg1 *= @s arg2
scoreboard players operation @s arg1 /= 100 num
scoreboard players operation @s hurt.total = @s arg1
# 真伤
scoreboard players operation @s hurt.total += @s hurt.real
# 更新显示挂留
execute @s[score_elemet.save.id_min=1] ~ ~ ~ execute @e[c=1,dy=2,tag=attribute.elemet] ~ ~ ~ function rpg:monster/hurt/elemet/ice/update_elemet
# 扣除生命值
scoreboard players operation @s health -= @s hurt.total
# 死亡
function rpg:monster/hurt/death/main if @s[score_health=0]
# 更新血量lv/护盾提示
execute @e[c=1,dy=2,tag=attribute.hp_armor] ~ ~ ~ function rpg:monster/hurt/elemet/update_hp_armor
# 数显
scoreboard players operation @s arg1 = @s hurt.total
function sys:api/数显/main
# 玩家标题更新
# 玩家返回指向事件
title @a[score_temp.uid.match_min=0,score_temp.uid.match=0] title [" §7◈ §e§l⚶ §7◈"]
title @a[score_temp.uid.match_min=0,score_temp.uid.match=0] subtitle ["§7⌯⌯◈ "," §a۞",{"score":{"objective":"armor.all.now","name":"@s"},"color":"yellow"}," | ",{"score":{"objective":"armor.all.max","name":"@s"},"color":"aqua"},"  §c♥",{"score":{"objective":"health","name":"@s"},"color":"yellow"}," -",{"score":{"objective":"hurt.total","name":"@s"},"color":"red"}," | ",{"score":{"objective":"health.max","name":"@s"},"color":"aqua"}," §7◈⌯⌯"]
scoreboard players tag @a[score_temp.uid.match_min=0,score_temp.uid.match=0] remove self.atk
scoreboard players reset @a[score_temp.uid.match_min=0,score_temp.uid.match=0] temp.uid.match
# 玩家返回指向事件
# reset
scoreboard players reset @s hurt.temp
scoreboard players reset @s hurt.total
scoreboard players reset @s hurt.ice
scoreboard players reset @s hurt.thunder
scoreboard players reset @s arg1
scoreboard players reset @s arg2
scoreboard players reset @s return
scoreboard players reset @s hurt.armor
scoreboard players reset @s hurt.real
scoreboard players reset @s elemet.save.id
effect @s 24 1 0 true