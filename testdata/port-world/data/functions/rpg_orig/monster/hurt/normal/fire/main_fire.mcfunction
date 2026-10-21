## 怪物受火元素攻击主接口
    function rpg:monster/hurt/normal/global_start
    # 受击反馈
    particle lava ~ ~0.5 ~ 0 0 0 0.25 5 normal @a[r=12]
    particle explode ~ ~0.2 ~ 0 0 0 0.25 3 normal @a[r=12]
    # 玩家返回指向事件
    execute @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ~ ~ ~ function rpg:monster/hurt/normal/fire/backToPlayer
    # 整合伤害到总计
        scoreboard players operation @s hurt.total += @s hurt.fire
    # 护盾阶段判断[结果——>hurt.temp]
        scoreboard players operation @s armor.all.now -= @s hurt.armor
        scoreboard players operation @s arg1 = @s armor.all.now
        scoreboard players operation @s arg1 *= 100 num
        scoreboard players operation @s arg1 /= @s armor.all.max
        scoreboard players set @s[score_arg1_min=80] hurt.temp 80
        scoreboard players set @s[score_arg1_min=40,score_arg1=79] hurt.temp 50
        scoreboard players set @s[score_arg1_min=0,score_arg1=39] hurt.temp 20
    # 额外削弱本系伤害【护盾+10%】
        scoreboard players add @s[tag=water] hurt.temp 10
        scoreboard players add @s[tag=fire] hurt.temp 5
    # 通用尾调
        function rpg:monster/hurt/normal/global_end
        scoreboard players reset @s hurt.fire
#