teleport @e[tag=snf,rm=10,r=15] ~ ~ ~2 
execute @e[tag=snf,type=!player] ~ ~ ~ scoreboard players add @e[tag=sum.snf,r=15] sum.nearByMob 1
execute @s[score_sum.nearByPlayer=0] ~ ~ ~ tp @e[tag=snf,r=5] ~ ~-1000 ~

execute @s[score_sum.nearByPlayer_min=1,score_sum.nearByMob=9] ~2 ~ ~ function rpg:sum/mob/1c/水匿浮/0
execute @s[score_sum.nearByPlayer_min=1,score_sum.nearByMob=9] ~ ~ ~5 function rpg:sum/mob/1c/水匿浮/0


particle bubble ~0.052357 ~-1 ~-2.999543 -0.034905 -0.0 1.999695 0.2 10
particle bubble ~1.075104 ~-1 ~-2.800741 -0.716736 -0.0 1.867161 0.2 10
particle bubble ~1.968177 ~-1 ~-2.264129 -1.312118 -0.0 1.509419 0.2 10
particle bubble ~2.623859 ~-1 ~-1.454429 -1.749239 -0.0 0.969619 0.2 10
particle bubble ~2.963065 ~-1 ~-0.469303 -1.975377 -0.0 0.312869 0.2 10
particle bubble ~2.944882 ~-1 ~0.572427 -1.963254 -0.0 -0.381618 0.2 10
particle bubble ~2.571502 ~-1 ~1.545114 -1.714335 -0.0 -1.030076 0.2 10
particle bubble ~1.887961 ~-1 ~2.331438 -1.258641 -0.0 -1.554292 0.2 10
particle bubble ~-1.968177 ~-1 ~2.264129 1.312118 -0.0 -1.509419 0.2 10
particle bubble ~-2.623859 ~-1 ~1.454429 1.749239 -0.0 -0.969619 0.2 10
particle bubble ~-2.963065 ~-1 ~0.469303 1.975377 -0.0 -0.312869 0.2 10
particle bubble ~-2.944882 ~-1 ~-0.572427 1.963254 -0.0 0.381618 0.2 10
particle bubble ~-2.571502 ~-1 ~-1.545114 1.714335 -0.0 1.030076 0.2 10
particle bubble ~-1.887961 ~-1 ~-2.331438 1.258641 -0.0 1.554292 0.2 10
particle bubble ~-0.976704 ~-1 ~-2.836556 0.651136 -0.0 1.891037 0.2 10

scoreboard players set @s sum.time 0
scoreboard players set @s sum.nearByMob 0