##本方法用于找出 当前记分板中分数最低的玩家
## 注意 使用前须将所有玩家的arg1 赋值为待比较记分板
##最高分数的玩家会拥有 最低分数 的tag

##准备部分
execute @a ~ ~ ~ function sys:math/寻找/ScoreNotFound unless @s[score_arg1_min=-2147483648]
scoreboard players tag @a[tag=最低分数] remove 最低分数
scoreboard players set num-SearchMin math.operation1 2147483647
##运算部分
execute @a ~ ~ ~ scoreboard players operation num-SearchMin math.operation1 < @s arg1
execute @a ~ ~ ~ scoreboard players operation @s return = @s arg1
execute @a ~ ~ ~ scoreboard players operation @s return -= num-SearchMin math.operation1
##运算结果
scoreboard players tag @a[score_return_min=0,score_return=0] add 最低分数
##重置
scoreboard players reset @a arg1
scoreboard players reset @a return
scoreboard players reset num-SearchMin math.operation1
