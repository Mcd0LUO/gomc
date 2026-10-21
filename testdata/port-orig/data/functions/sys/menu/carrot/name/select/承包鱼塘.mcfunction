scoreboard players operation @s health.max -= @s name.+HpMax
scoreboard players operation @s mp.max -= @s name.+MpMax

scoreboard players set @s name.+HpMax 60
scoreboard players set @s name.+MpMax 30
scoreboard teams join 承包鱼塘 @s

scoreboard players operation @s health.max += @s name.+HpMax
scoreboard players operation @s mp.max += @s name.+MpMax