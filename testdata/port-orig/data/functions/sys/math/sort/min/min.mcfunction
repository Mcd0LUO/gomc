# 方法用于从小到大sort玩家分数
# 注意 使用前须 使待比较玩家的arg1 = 待比较记分板

# 初始化
scoreboard players set num-SearchMax math.operation1 2147483647
# 玩家分数小于num-SearchMax则赋值num-SearchMin=当前玩家分数
execute @a[tag=!已参与] ~ ~ ~ scoreboard players operation num-SearchMax math.operation1 < @s arg1
execute @a[tag=!已参与] ~ ~ ~ scoreboard players operation @s math.operation1 = @s arg1
# 判断最低分数的玩家
execute @a[tag=!已参与] ~ ~ ~ scoreboard players operation @s math.operation1 -= num-SearchMax math.operation1
# 添加标签并加入已参与的列表
scoreboard players tag @a[score_math.operation1_min=0,score_math.operation1=0] add 最低分数
scoreboard players tag @a[score_math.operation1_min=0,score_math.operation1=0] add 已参与
# 重置
scoreboard players reset @a math.operation1
scoreboard players reset num-SearchMin math.operation1