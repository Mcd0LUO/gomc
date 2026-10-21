
summon minecraft:armor_stand ~ ~ ~ {Tags:["skill.mob","skill.岩狱之路.new","skill.岩狱之路","skill.岩狱之路.2"],Invulnerable:1b,PersistenceRequired:1b,NoBasePlate:1b,ShowArms:1b,DisabledSlots:65793,Invisible:1b,ArmorItems:[{},{},{},{id:"minecraft:magma",Count:1b,Damage:0s}],Pose:{Head:[-180f,0f,0f],RightArm:[-50f,0f,0f]},HandItems:[{id:"minecraft:magma",Count:1b,Damage:0s}]}
tp @e[c=1,tag=skill.岩狱之路.new] @s
tp @e[c=1,tag=skill.岩狱之路.new] ~ ~ ~


scoreboard players operation @e[c=1,tag=skill.岩狱之路.new] uid = @s uid
scoreboard players operation @e[c=1,tag=skill.岩狱之路.new] damage.fire = @s damage.fire
entitydata @e[c=1,tag=skill.岩狱之路.new] {Motion:[0.0,0.8,0.0]}
scoreboard players tag @e[c=1,tag=skill.岩狱之路.new] remove skill.岩狱之路.new


function sys:api/视角选择器/horizontal/tp/1/main
particle flame ~ ~1 ~ 0.1 0 0.1 0.5 3