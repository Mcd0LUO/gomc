execute @s[score_skill.step_min=1,score_skill.step=1] ~ ~ ~ summon minecraft:armor_stand ~1.3 ~ ~-0.5 {Tags:["skill.mob","skill.寒星环刃5","skill.寒星环刃.5.new","skill.寒星环刃","skill.fly"],Invulnerable:1b,Marker:1b,ShowArms:1,PersistenceRequired:1b,NoGravity:1b,HandItems:[{id:"minecraft:iron_sword",Count:1b,Damage:0s,tag:{ench:[{id:34s,lvl:1s}]}},{}],Invisible:1,Rotation:[0f],Pose:{RightArm:[0f,0f,0f]}}
execute @s[score_skill.step_min=1,score_skill.step=1] ~ ~ ~ kill @e[c=1,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0]

execute @s[score_skill.step_min=2,score_skill.step=2] ~ ~ ~ summon minecraft:armor_stand ~1.3 ~ ~-0.5 {Tags:["skill.mob","skill.寒星环刃5","skill.寒星环刃.5.new","skill.寒星环刃","skill.fly"],Invulnerable:1b,Marker:1b,ShowArms:1,PersistenceRequired:1b,NoGravity:1b,HandItems:[{id:"minecraft:iron_sword",Count:1b,Damage:0s,tag:{ench:[{id:34s,lvl:1s}]}},{}],Invisible:1,Rotation:[0f],Pose:{RightArm:[0f,0f,0f]}}
execute @s[score_skill.step_min=2,score_skill.step=2] ~ ~ ~ kill @e[c=1,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0]

execute @s[score_skill.step_min=3,score_skill.step=3] ~ ~ ~ summon minecraft:armor_stand ~1.3 ~ ~-0.5 {Tags:["skill.mob","skill.寒星环刃5","skill.寒星环刃.5.new","skill.寒星环刃","skill.fly"],Invulnerable:1b,Marker:1b,ShowArms:1,PersistenceRequired:1b,NoGravity:1b,HandItems:[{id:"minecraft:iron_sword",Count:1b,Damage:0s,tag:{ench:[{id:34s,lvl:1s}]}},{}],Invisible:1,Rotation:[0f],Pose:{RightArm:[0f,0f,0f]}}
execute @s[score_skill.step_min=3,score_skill.step=3] ~ ~ ~ kill @e[c=1,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0]

execute @s[score_skill.step_min=4,score_skill.step=4] ~ ~ ~ summon minecraft:armor_stand ~1.3 ~ ~-0.5 {Tags:["skill.mob","skill.寒星环刃5","skill.寒星环刃.5.new","skill.寒星环刃","skill.fly"],Invulnerable:1b,Marker:1b,ShowArms:1,PersistenceRequired:1b,NoGravity:1b,HandItems:[{id:"minecraft:iron_sword",Count:1b,Damage:0s,tag:{ench:[{id:34s,lvl:1s}]}},{}],Invisible:1,Rotation:[0f],Pose:{RightArm:[0f,0f,0f]}}
execute @s[score_skill.step_min=4,score_skill.step=4] ~ ~ ~ kill @e[c=1,tag=skill.寒星环刃,score_skill.temp_min=0,score_skill.temp=0]


tp @e[c=1,tag=skill.寒星环刃5] @s
tp @e[c=1,tag=skill.寒星环刃5] ~ ~0.5 ~
scoreboard players operation @e[c=1,tag=skill.寒星环刃5] uid = @s uid
scoreboard players set @e[c=1,tag=skill.寒星环刃5] damage.ice 80
scoreboard players operation @e[c=1,tag=skill.寒星环刃5] damage.ice += @s elemet.ice_lv
scoreboard players set @e[c=1,tag=skill.寒星环刃5] elemet.save.id 3
scoreboard players set @e[c=1,tag=skill.寒星环刃5] damage.real 3

scoreboard players reset @s[score_skill.step_min=4] skill.id
scoreboard players reset @s[score_skill.step_min=4] skill.time.ice
scoreboard players reset @s[score_skill.step_min=4] skill.step