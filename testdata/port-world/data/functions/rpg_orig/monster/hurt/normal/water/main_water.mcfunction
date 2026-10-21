## 怪物受水元素攻击主接口
    function rpg:monster/hurt/normal/global_start

    # 玩家返回指向事件
    execute @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ~ ~ ~ function rpg:monster/hurt/normal/water/backToPlayer 
    # 整合伤害到总计
    scoreboard players operation @s hurt.total += @s hurt.water
    # 额外削弱本系伤害【护盾%】
    scoreboard players add @s[tag=water] hurt.temp 5
    scoreboard players remove @s[tag=fire] hurt.temp 10
    # 获取护盾减伤阶段
    function rpg:monster/hurt/api/armor_lv/get_armor_term
    # 通用尾调
    function rpg:monster/hurt/normal/global_end
    scoreboard players reset @s hurt.water

#