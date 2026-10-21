effect @s 2 4 1 true
scoreboard players operation @s health -= #2 ggm.c
scoreboard players operation @s t5.ggm = @s health
scoreboard players operation @s t5.ggm > #0 ggm.c
scoreboard players operation @s health = @s t5.ggm
scoreboard players set @s debuff.armor 30
title @s title [""]
title @s subtitle ["§b寒冷值ꇻ",{"score":{"objective":"bar.cold","name":"@s"},"color":"red"},"|",{"score":{"objective":"bar.cold.max","name":"@s"},"color":"red"}," §eDebuff:§r减速/HP/法抗                            "]
