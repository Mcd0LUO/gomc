##本方法用于从大到小sort玩家
## 注意 使用前须将所有玩家的arg1 赋值为待比较记分板

# 初始化
scoreboard players set num-SearchMax math.operation1 -2147483648
# 玩家分数大于num-SearchMax则赋值num-SearchMax=当前玩家分数
execute @a[tag=!已参与] ~ ~ ~ scoreboard players operation num-SearchMax math.operation1 > @s arg1
execute @a[tag=!已参与] ~ ~ ~ scoreboard players operation @s math.operation1 = @s arg1
# 判断最高分数的玩家
execute @a[tag=!已参与] ~ ~ ~ scoreboard players operation @s math.operation1 -= num-SearchMax math.operation1
# 添加标签并加入已参与的列表
scoreboard players tag @a[score_math.operation1_min=0,score_math.operation1=0] add 最高分数
scoreboard players tag @a[score_math.operation1_min=0,score_math.operation1=0] add 已参与
# 重置
scoreboard players reset @a math.operation1
scoreboard players reset num-SearchMax math.operation1