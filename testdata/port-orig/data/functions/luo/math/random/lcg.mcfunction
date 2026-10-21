# 使用前需 赋值arg1(min) arg2(max) 之后调用本函数
# 产生随机数
scoreboard players operation @s return = seed math
execute @s[score_return=-1] ~ ~ ~ scoreboard players set seed math 12345
scoreboard players operation @s return *= a math
scoreboard players operation @s return += c math
scoreboard players operation @s return %= m math
# 映射到范围
scoreboard players operation @s arg2 -= @s arg1
scoreboard players add @s arg2 1
scoreboard players operation @s return %= @s arg2
# 绝对值
scoreboard players operation @s[score_return=-1] return *= math.-1 math
scoreboard players operation @s return += @s arg1
# 重置
scoreboard players operation seed math += c math
scoreboard players reset @s arg1
scoreboard players reset @s arg2