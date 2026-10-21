effect @s 2 4 1 true
scoreboard players operation @s health -= #1 ggm.c
scoreboard players operation @s t4.ggm = @s health
scoreboard players operation @s t4.ggm > #0 ggm.c
scoreboard players operation @s health = @s t4.ggm
title @s subtitle ["§b寒冷值ꇻ",{"score":{"objective":"bar.cold","name":"@s"},"color":"red"},"|",{"score":{"objective":"bar.cold.max","name":"@s"},"color":"red"}," §eDebuff:§r减速/HP                               "]
