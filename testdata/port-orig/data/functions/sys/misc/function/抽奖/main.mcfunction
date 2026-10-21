# execute @s[score_dq_min=180] ~ ~ ~ function sys:misc/function/抽奖/单抽 if @e[c=1,tag=单抽,r=3]
# execute @s[score_dq=179] ~ ~ ~ function sys:misc/function/抽奖/fail if @e[c=1,tag=单抽,r=3]
# execute @s[score_dq_min=1800] ~ ~ ~ function sys:misc/function/抽奖/十连抽 if @e[c=1,tag=十连抽,r=3]
# execute @s[score_dq=1799] ~ ~ ~ function sys:misc/function/抽奖/fail if @e[c=1,tag=十连抽,r=3]

scoreboard players reset @s talk
tellraw @s ["§r[§e埃瑞涅的馈赠§r]:§7很抱歉，奖池暂未开启哦"]