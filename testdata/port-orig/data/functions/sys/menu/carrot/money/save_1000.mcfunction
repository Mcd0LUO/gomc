stats entity @s set AffectedItems @s stats.item
scoreboard players set @s stats.item 0
clear @s paper 0 -1 {Tags:["货币","末法点","1000"]}
scoreboard players operation @s stats.item *= 1000 num
scoreboard players operation @s money += @s stats.item
tellraw @s ["§r[§eBANK§r]:§7存入虚拟末法点 ",{"score":{"objective":"stats.item","name":"@s"},"color":"red"}]
scoreboard players reset @s stats.item
stats entity @s clear AffectedItems @s