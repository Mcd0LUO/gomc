##本方法用于找出 当前记分板中分数最高的玩家
## 注意 使用前须将所有玩家的arg1 赋值为待比较记分板
##最高分数的玩家会拥有 最高分数 的tag

##准备部分
execute @a ~ ~ ~ function sys:math/寻找/ScoreNotFound unless @s[score_arg1_min=-2147483648]
scoreboard players tag @a[tag=最高分数] remove 最高分数
scoreboard players set num-SearchMax math.operation1 -2147483648
##运算部分
execute @a[tag=!ScoreNotFound] ~ ~ ~ scoreboard players operation num-SearchMax math.operation1 > @s arg1
execute @a[tag=!ScoreNotFound] ~ ~ ~ scoreboard players operation @s return = @s arg1
execute @a[tag=!ScoreNotFound] ~ ~ ~ scoreboard players operation @s return -= num-SearchMax math.operation1
##运算结果
scoreboard players tag @a[score_return_min=0,score_return=0] add 最高分数
##重置
scoreboard players tag @a[tag=ScoreNotFound] remove ScoreNotFound
scoreboard players reset @a arg1
scoreboard players reset @a return
scoreboard players reset num-SearchMax math.operation1
