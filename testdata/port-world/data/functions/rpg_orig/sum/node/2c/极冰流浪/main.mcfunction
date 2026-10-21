teleport @e[tag=极冰流浪,rm=10,r=15] ~ ~ ~2 
execute @e[tag=极冰流浪] ~ ~ ~ scoreboard players add @e[tag=sum.极冰流浪,r=15] sum.nearByMob 1
execute @s[score_sum.nearByPlayer=0] ~ ~ ~ tp @e[tag=极冰流浪,r=5] ~ ~-1000 ~

execute @s[score_sum.nearByPlayer_min=1,score_sum.nearByMob=9] ~2 ~ ~ function rpg:sum/mob/2c/极冰流浪/0

execute @s[score_sum.nearByPlayer_min=1,score_sum.nearByMob=9] ~ ~ ~4 function rpg:sum/mob/2c/极冰流浪/0


particle iconcrack ~ ~ ~ 0 0 0 1 50 normal @a[r=10] 264

scoreboard players set @s sum.time 0
scoreboard players set @s nearByMob 0