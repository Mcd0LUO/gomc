scoreboard players remove @s angry 120



execute @e[tag=mob.target,r=4] ~ ~ ~ particle lava ~ ~1 ~ 0.3 0 0.3 1 5
execute @e[tag=mob.target,r=4] ~ ~ ~ /particle fireworksSpark ~ ~1 ~ 0.3 0.3 0.3 0.4 4 normal
scoreboard players set @e[r=3,tag=mob.target] hurt.fire 20
execute @e[tag=mob.target,r=3] ~ ~ ~ function rpg:monster/hurt/main 