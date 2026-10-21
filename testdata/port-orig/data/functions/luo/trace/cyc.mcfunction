    summon armor_stand ~1 ~ ~ {Tags:["api","api.trace","api.new"],NoGravity:1b,Marker:1b,Invisible:1b}
    summon armor_stand ~-1 ~ ~ {Tags:["api","api.trace","api.new"],NoGravity:1b,Marker:1b,Invisible:1b}
    summon armor_stand ~ ~ ~1 {Tags:["api","api.trace","api.new"],NoGravity:1b,Marker:1b,Invisible:1b}
    summon armor_stand ~ ~ ~-1 {Tags:["api","api.trace","api.new"],NoGravity:1b,Marker:1b,Invisible:1b}
    summon armor_stand ~ ~1 ~ {Tags:["api","api.trace","api.new"],NoGravity:1b,Marker:1b,Invisible:1b}
    summon armor_stand ~ ~-1 ~ {Tags:["api","api.trace","api.new"],NoGravity:1b,Marker:1b,Invisible:1b}
    scoreboard players remove @s arg1 1
    scoreboard players operation @e[r=2,tag=api.new] M.UID = @s M.UID 
    execute @s ~ ~ ~ function luo:trace/Exception unless @e[r=12,tag=api.trace_target]
    execute @e[r=12,tag=api.trace_target] ~ ~ ~ execute @e[r=12,c=1,type=armor_stand,tag=api.new] ~ ~ ~ tp @e[r=2,c=1,tag=api.trace_main] @s
    # execute @s[score_arg1_min=1] ~ ~ ~ function luo:trace/cyc unless @e[r=1,c=1,tag=api.trace_target]