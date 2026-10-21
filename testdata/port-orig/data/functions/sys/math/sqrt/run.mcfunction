##调用本函数即可开平方  (牛顿迭代法)
##使用前需赋值 待开平方数 arg1
function sys:math/开平方/ini unless @s[score_math.operation1_min=-2147483647]
#估计值
scoreboard players operation @s math.operation2 = @s 运算值3
scoreboard players operation @s math.operation2 /= @s math.operation1
scoreboard players operation @s math.operation1 += @s math.operation2
scoreboard players operation @s math.operation1 /= 2 num
# 收敛
scoreboard players operation @s arg1 = @s math.operation1
scoreboard players operation @s num = @s math.operation1
scoreboard players operation @s num -= @s math.operation2
function sys:math/开平方/开平方 unless @s[score_num_min=-1,score_num=1]
function sys:math/开平方/end if @s[score_num_min=-1,score_num=1]