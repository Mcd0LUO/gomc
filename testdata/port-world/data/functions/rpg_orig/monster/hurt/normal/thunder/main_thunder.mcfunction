# 怪物受火元素攻击主接口
# 受击反馈
particle smoke ~ ~ ~ 0 0 0 0.25 20 normal
particle witchMagic ~ ~2.2 ~ 0 0 0 0.25 10 normal
particle endRod ~ ~1 ~ 0 0 0 0.1 5 normal
# 获取攻击玩家
scoreboard players operation @a[tag=self.atk,r=12] temp.uid.match = @s uid
execute @a[tag=self.atk,r=12] ~ ~ ~ scoreboard players operation @s temp.uid.match -= @s uid
# 玩家返回指向事件
execute @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ~ ~ ~ function rpg:monster/hurt/normal/thunder/backToPlayer
# 运算受到的伤害
# 整合伤害到总计
scoreboard players operation @s hurt.total += @s hurt.thunder
# 随机减护盾
scoreboard players operation @s arg1 = @r data.gameTime
scoreboard players operation @s arg1 %= 2 num
scoreboard players operation @s armor.all.now -= @s arg1
# 护盾阶段判断[结果——>hurt.temp]
scoreboard players operation @s armor.all.now -= @s hurt.armor
scoreboard players operation @s arg1 = @s armor.all.now
scoreboard players operation @s arg1 *= 100 num
scoreboard players operation @s arg1 /= @s armor.all.max
scoreboard players set @s[score_arg1_min=80] hurt.temp 80
scoreboard players set @s[score_arg1_min=40,score_arg1=79] hurt.temp 50
scoreboard players set @s[score_arg1_min=0,score_arg1=39] hurt.temp 20
# 额外削弱本系伤害【护盾+10%】
scoreboard players remove @s[tag=water] hurt.temp 10
scoreboard players add @s[tag=thunder] hurt.temp 5
# 最终减伤 arg1为结果
scoreboard players set @s arg1 0
scoreboard players operation @s arg1 = @s hurt.total
scoreboard players set @s arg2 100
scoreboard players operation @s arg2 -= @s hurt.temp
scoreboard players operation @s arg1 *= @s arg2
scoreboard players operation @s arg1 /= 100 num
scoreboard players operation @s hurt.total = @s arg1
# 扣除生命值
scoreboard players operation @s health -= @s hurt.total
function rpg:monster/hurt/death/main if @s[score_health=0]
# 更新血量lv/护盾提示
execute @e[c=1,dy=1,tag=attribute.hp_armor] ~ ~ ~ function rpg:monster/hurt/elemet/update_hp_armor
# 玩家标题更新
title @a[score_temp.uid.match_min=0,score_temp.uid.match=0] title [" §7◈ §e§l⚶ §7◈"]
title @a[score_temp.uid.match_min=0,score_temp.uid.match=0] subtitle ["§7⌯⌯◈ "," §a۞",{"score":{"objective":"armor.all.now","name":"@s"},"color":"yellow"}," | ",{"score":{"objective":"armor.all.max","name":"@s"},"color":"aqua"},"  §c♥",{"score":{"objective":"health","name":"@s"},"color":"yellow"}," -",{"score":{"objective":"hurt.total","name":"@s"},"color":"red"}," | ",{"score":{"objective":"health.max","name":"@s"},"color":"aqua"}," §7◈⌯⌯"]
scoreboard players tag @a[score_temp.uid.match_min=0,score_temp.uid.match=0] remove self.atk
scoreboard players reset @a[score_temp.uid.match_min=0,score_temp.uid.match=0] temp.uid.match
# reset
scoreboard players reset @s hurt.temp
scoreboard players reset @s hurt.total
scoreboard players reset @s hurt.thunder
scoreboard players reset @s arg1
scoreboard players reset @s arg2
scoreboard players reset @s return
scoreboard players reset @s hurt.armor
scoreboard players tag @s remove mob.hurt
effect @s 24 1 0 true