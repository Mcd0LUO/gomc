function sys:skill/press/ice/寒星环刃/ini unless @s[score_skill.time.ice_min=0]
scoreboard players add @s skill.time.ice 1
# 匹配uid
scoreboard players operation @e[tag=skill.寒星环刃,r=10] skill.temp = @s uid
execute @e[tag=skill.寒星环刃,r=10] ~ ~ ~ scoreboard players operation @s skill.temp -= @s uid
execute @e[c=4,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0] ~ ~1 ~ particle endRod ~ ~1 ~ 0 -1 0 0.5 0
#execute @e[c=4,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0] ~ ~1 ~ particle iconcrack ~ ~1 ~ 0 0.1 0 0.1 0 normal @a[r=6] 264

execute @s[score_skill.time.ice_min=1,score_skill.time.ice=1] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~0.0 ~0.1 ~-2.0 ~ ~
execute @s[score_skill.time.ice_min=2,score_skill.time.ice=2] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~1.285575 ~0.1 ~-1.532089 ~ ~
execute @s[score_skill.time.ice_min=3,score_skill.time.ice=3] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~1.969616 ~0.1 ~-0.347296 ~ ~
execute @s[score_skill.time.ice_min=4,score_skill.time.ice=4] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~1.732051 ~0.1 ~1.0 ~ ~
execute @s[score_skill.time.ice_min=5,score_skill.time.ice=5] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~0.68404 ~0.1 ~1.879385 ~ ~
execute @s[score_skill.time.ice_min=6,score_skill.time.ice=6] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~-0.68404 ~0.1 ~1.879385 ~ ~
execute @s[score_skill.time.ice_min=7,score_skill.time.ice=7] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~-1.732051 ~0.1 ~1.0 ~ ~
execute @s[score_skill.time.ice_min=8,score_skill.time.ice=8] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~-1.969616 ~0.1 ~-0.347296 ~ ~
execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.1,score_skill.temp_min=0,score_skill.temp=0] ~-1.285575 ~0.1 ~-1.532089 ~ ~

execute @s[score_skill.time.ice_min=1,score_skill.time.ice=1] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~2.0 ~0.1 ~0.0 ~ ~
execute @s[score_skill.time.ice_min=2,score_skill.time.ice=2] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~1.532089 ~0.1 ~1.285575 ~ ~
execute @s[score_skill.time.ice_min=3,score_skill.time.ice=3] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~0.347296 ~0.1 ~1.969616 ~ ~
execute @s[score_skill.time.ice_min=4,score_skill.time.ice=4] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~-1.0 ~0.1 ~1.732051 ~ ~
execute @s[score_skill.time.ice_min=5,score_skill.time.ice=5] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~-1.879385 ~0.1 ~0.68404 ~ ~
execute @s[score_skill.time.ice_min=6,score_skill.time.ice=6] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~-1.879385 ~0.1 ~-0.68404 ~ ~
execute @s[score_skill.time.ice_min=7,score_skill.time.ice=7] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~-1.0 ~0.1 ~-1.732051 ~ ~
execute @s[score_skill.time.ice_min=8,score_skill.time.ice=8] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~0.347296 ~0.1 ~-1.969616 ~ ~
execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.2,score_skill.temp_min=0,score_skill.temp=0] ~1.532089 ~0.1 ~-1.285575 ~ ~


execute @s[score_skill.time.ice_min=1,score_skill.time.ice=1] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~-0.0 ~0.1 ~2.0 ~ ~
execute @s[score_skill.time.ice_min=2,score_skill.time.ice=2] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~-1.285575 ~0.1 ~1.532089 ~ ~
execute @s[score_skill.time.ice_min=3,score_skill.time.ice=3] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~-1.969616 ~0.1 ~0.347296 ~ ~
execute @s[score_skill.time.ice_min=4,score_skill.time.ice=4] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~-1.732051 ~0.1 ~-1.0 ~ ~
execute @s[score_skill.time.ice_min=5,score_skill.time.ice=5] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~-0.68404 ~0.1 ~-1.879385 ~ ~
execute @s[score_skill.time.ice_min=6,score_skill.time.ice=6] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~0.68404 ~0.1 ~-1.879385 ~ ~
execute @s[score_skill.time.ice_min=7,score_skill.time.ice=7] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~1.732051 ~0.1 ~-1.0 ~ ~
execute @s[score_skill.time.ice_min=8,score_skill.time.ice=8] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~1.969616 ~0.1 ~0.347296 ~ ~
execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.3,score_skill.temp_min=0,score_skill.temp=0] ~1.285575 ~0.1 ~1.532089 ~ ~

execute @s[score_skill.time.ice_min=1,score_skill.time.ice=1] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~-2.0 ~0.1 ~-0.0 ~ ~
execute @s[score_skill.time.ice_min=2,score_skill.time.ice=2] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~-1.532089 ~0.1 ~-1.285575 ~ ~
execute @s[score_skill.time.ice_min=3,score_skill.time.ice=3] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~-0.347296 ~0.1 ~-1.969616 ~ ~
execute @s[score_skill.time.ice_min=4,score_skill.time.ice=4] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~1.0 ~0.1 ~-1.732051 ~ ~
execute @s[score_skill.time.ice_min=5,score_skill.time.ice=5] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~1.879385 ~0.1 ~-0.68404 ~ ~
execute @s[score_skill.time.ice_min=6,score_skill.time.ice=6] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~1.879385 ~0.1 ~0.68404 ~ ~
execute @s[score_skill.time.ice_min=7,score_skill.time.ice=7] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~1.0 ~0.1 ~1.732051 ~ ~
execute @s[score_skill.time.ice_min=8,score_skill.time.ice=8] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~-0.347296 ~0.1 ~1.969616 ~ ~
execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ teleport @e[c=4,tag=skill.寒星环刃.4,score_skill.temp_min=0,score_skill.temp=0] ~-1.532089 ~0.1 ~1.285575 ~ ~

execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ scoreboard players add @s skill.step 1
execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ function sys:skill/press/ice/寒星环刃/shot
execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ scoreboard players set @s skill.time.ice 0

#execute @s[score_skill.time.ice_min=9,score_skill.time.ice=9] ~ ~ ~ kill @e[c=4,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0]
#execute @e[c=3,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0] ~ ~ ~ execute @e[c=2,r=2,tag=mob.asg] ~ ~ ~ 