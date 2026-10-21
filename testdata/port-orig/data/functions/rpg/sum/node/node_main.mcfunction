scoreboard players add @s sum.time 1
# 计算半径6周围存在的玩家
scoreboard players set @s sum.nearByPlayer 0
execute @a[r=10] ~ ~ ~ scoreboard players add @e[c=1,tag=sum,type=armor_stand] sum.nearByPlayer 1
# select
function rpg:sum/node/1c/火布鲁/main if @s[tag=sum.hbl,score_sum.time_min=11]
function rpg:sum/node/1c/水匿浮/main if @s[tag=sum.snf,score_sum.time_min=11]
function rpg:sum/node/1c/蓝露子/main if @s[tag=sum.llz,score_sum.time_min=11]
function rpg:sum/node/2c/极冰流浪/main if @s[tag=sum.极冰流浪,score_sum.time_min=11]

# reset
scoreboard players set @s sum.nearByMob 0