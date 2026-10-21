scoreboard players set @s skill.time.ice 0
scoreboard players set @s cd.i 20

summon minecraft:armor_stand ~ ~ ~ {Tags:["skill.mob","skill.岩狱之路.new","skill.岩狱之路","skill.岩狱之路.1"],Invulnerable:1b,PersistenceRequired:1b,NoBasePlate:1b,DisabledSlots:65793,NoGravity:1b,Invisible:1b,ArmorItems:[{},{},{},{id:"minecraft:magma",Count:1b,Damage:0s}],Pose:{Head:[-180f,0f,0f]}}
tp @e[c=1,tag=skill.岩狱之路.new] @s
tp @e[c=1,tag=skill.岩狱之路.new] ~ ~-1 ~

scoreboard players operation @e[c=1,tag=skill.岩狱之路.new] uid = @s uid
scoreboard players set @e[c=1,tag=skill.岩狱之路.new] damage.fire 30
scoreboard players operation @e[c=1,tag=skill.岩狱之路.new] damage.fire += @s elemet.fire_lv
scoreboard players tag @e[c=1,tag=skill.岩狱之路.new] remove skill.岩狱之路.new
scoreboard players reset @s skill.step
effect @s minecraft:slowness 1 1 true