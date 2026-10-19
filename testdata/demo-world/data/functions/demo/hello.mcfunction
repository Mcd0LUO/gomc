# A small 1.12.2 style function.
scoreboard objectives add counter dummy
scoreboard players set Steve counter 1
scoreboard players add Steve counter 2
summon armor_stand ~ ~ ~
execute @e[type=armor_stand] ~ ~1 ~ setblock ~ ~ ~ stone
summon zombie ~ ~ ~
scoreboard players tag @e[type=zombie] add spawned
say demo function finished
