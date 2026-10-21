scoreboard players operation @s mp.now += #1 ggm.c
scoreboard players operation @s mp.now += @s mp.MoreRestore
scoreboard players operation @s t1.ggm = @s mp.now
scoreboard players operation @s t1.ggm < @s mp.max
scoreboard players operation @s mp.now = @s t1.ggm
