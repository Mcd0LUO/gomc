##本方法用于倒置传入的数值 例如传入 12345 那么将会输出 54321
scoreboard players set num-10 num 10
scoreboard players operation @s arg2 = @s arg1 
scoreboard players operation @s arg2 %= num-10 num
scoreboard players operation @s arg1 /= num-10 num

scoreboard players operation @s return *= num-10 num
scoreboard players operation @s return += @s arg2

function sys:math/reverse if @s[score_arg1_min=1]

scoreboard players reset @s arg1
scoreboard players reset @s arg2
scoreboard players reset num-10 num