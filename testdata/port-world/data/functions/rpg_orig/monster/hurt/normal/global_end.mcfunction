## 通用尾调函数
    # 最终减伤 arg1为最终伤害
    scoreboard players set @s arg1 0
    scoreboard players operation @s arg1 = @s hurt.total
    scoreboard players set @s arg2 100
    scoreboard players operation @s arg2 -= @s hurt.temp
    scoreboard players operation @s arg1 *= @s arg2
    scoreboard players operation @s arg1 /= 100 num
    scoreboard players operation @s hurt.total = @s arg1
    # 扣除生命值
    scoreboard players operation @s health -= @s hurt.total
    function rpg:player/hurt/death/main if @s[score_health=0]
    ## 数显
    scoreboard players operation @s arg1 = @s hurt.total
    function sys:api/数显/main
    # 更新血量lv/护盾提示
    scoreboard players remove @s armor.all.now 5
    execute @e[c=1,dy=2,tag=attribute.hp_armor] ~ ~ ~ function rpg:monster/hurt/api/ar_hp
    title @a[score_temp.uid.match_min=0,score_temp.uid.match=0] "§7◈ §e§l⚶ §7◈"
    title @a[score_temp.uid.match_min=0,score_temp.uid.match=0] subtitle ["§7⌯⌯◈ ","  §c♥",{"score":{"objective":"health","name":"@s"},"color":"yellow"}," -",{"score":{"objective":"hurt.total","name":"@s"},"color":"red"}," | ",{"score":{"objective":"health.max","name":"@s"},"color":"aqua"}," §7◈⌯⌯"]
    scoreboard players tag @a[score_temp.uid.match_min=0,score_temp.uid.match=0] remove self.atk
    scoreboard players reset @a[score_temp.uid.match_min=0,score_temp.uid.match=0] temp.uid.match

    # 清理
    scoreboard players reset @s hurt.temp
    scoreboard players reset @s hurt.total
    scoreboard players reset @s arg1
    scoreboard players reset @s arg2
    scoreboard players reset @s return
    scoreboard players reset @s hurt.armor
    scoreboard players tag @s remove mob.hurt
    effect @s 24 1 0 true

#