scoreboard players operation @s math.operation1 = @s money
scoreboard players remove @a math.operation1 1300

function sys:menu/chest/鱼钓之季/钓竿/专业_2 if @s[score_math.operation1_min=1300]

tellraw @s[score_math.operation1=1299] ["§r[§e鱼钓之季§r]:无效的操作，金币不足"]
scoreboard players remove @s[score_arg1_min=1300] money 1300
#
scoreboard players reset @s arg1
scoreboard players reset @s arg2
scoreboard players reset @s math.operation1