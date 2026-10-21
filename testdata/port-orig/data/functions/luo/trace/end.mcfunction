scoreboard players operation @e[r=12,tag=api.trace] temp1 = @s temp1 
execute @e[r=12,tag=api.trace] ~ ~ ~ scorebaordp players operation @s tem1 -= @s M.UID
kill @e[r=12,tag=api.trace,score_temp1_min=0,score_temp1=0]
scoreboard players reset @e[r=12,tag=api.trace] temp1 
kill @s
# say End