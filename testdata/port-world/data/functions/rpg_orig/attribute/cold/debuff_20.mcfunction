effect @s 2 4 1 true
scoreboard players remove @s health 1
scoreboard players operation @s health > num 0 num
title @s subtitle ["§b寒冷值ꇻ",{"score":{"objective":"bar.cold","name":"@s"},"color":"red"},"|",{"score":{"objective":"bar.cold.max","name":"@s"},"color":"red"}," §eDebuff:§r减速/HP                               "]