## 通用获取护盾阶段
    # 护盾阶段判断[结果——>hurt.temp]
    scoreboard players operation @s armor.all.now -= @s hurt.armor
    scoreboard players operation @s arg1 = @s armor.all.now
    scoreboard players operation @s arg1 *= 100 num
    scoreboard players operation @s arg1 /= @s armor.all.max
    scoreboard players set @s[score_arg1_min=80] hurt.temp 80
    scoreboard players set @s[score_arg1_min=40,score_arg1=79] hurt.temp 50
    scoreboard players set @s[score_arg1_min=0,score_arg1=39] hurt.temp 20

#