scoreboard players operation @s health.max -= @s name.+HpMax
scoreboard players operation @s mp.max -= @s name.+MpMax
scoreboard players operation @s health.restore -= @s name.+HpRestore

scoreboard players set @s name.+HpMax 100
scoreboard players set @s name.+MpMax 40
scoreboard players set @s name.+HpRestore 1
scoreboard teams join 承包鱼塘 @s

scoreboard players operation @s health.max += @s name.+HpMax
scoreboard players operation @s mp.max += @s name.+MpMax
scoreboard players operation @s health.restore += @s name.+HpRestore