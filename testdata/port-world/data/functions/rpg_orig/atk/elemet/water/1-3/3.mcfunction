particle endRod ~0.052357 ~0.2 ~-2.999543 -0.052357 0 2.999543 0.2 0
particle endRod ~0.572427 ~0.2 ~-2.944882 -0.572427 0 2.944882 0.2 0
particle endRod ~1.075104 ~0.2 ~-2.800741 -1.075104 0 2.800741 0.2 0
particle endRod ~1.545114 ~0.2 ~-2.571502 -1.545114 0 2.571502 0.2 0
particle endRod ~1.968177 ~0.2 ~-2.264129 -1.968177 0 2.264129 0.2 0
particle endRod ~2.331438 ~0.2 ~-1.887961 -2.331438 0 1.887961 0.2 0
particle endRod ~2.623859 ~0.2 ~-1.454429 -2.623859 0 1.454429 0.2 0
particle endRod ~2.836556 ~0.2 ~-0.976704 -2.836556 0 0.976704 0.2 0
particle endRod ~2.963065 ~0.2 ~-0.469303 -2.963065 0 0.469303 0.2 0
particle endRod ~2.999543 ~0.2 ~0.052357 -2.999543 0 -0.052357 0.2 0
particle endRod ~2.944882 ~0.2 ~0.572427 -2.944882 0 -0.572427 0.2 0
particle endRod ~2.800741 ~0.2 ~1.075104 -2.800741 0 -1.075104 0.2 0
particle endRod ~2.571502 ~0.2 ~1.545114 -2.571502 0 -1.545114 0.2 0
particle endRod ~2.264129 ~0.2 ~1.968177 -2.264129 0 -1.968177 0.2 0
particle endRod ~1.887961 ~0.2 ~2.331438 -1.887961 0 -2.331438 0.2 0
particle endRod ~1.454429 ~0.2 ~2.623859 -1.454429 0 -2.623859 0.2 0
particle endRod ~0.976704 ~0.2 ~2.836556 -0.976704 0 -2.836556 0.2 0
particle endRod ~0.469303 ~0.2 ~2.963065 -0.469303 0 -2.963065 0.2 0

particle splash ~1.5 ~ ~ 0 2 0 0.4 35 normal
particle splash ~-1.5 ~ ~ 0 2 0 0.4 35 normal
particle splash ~ ~ ~1 0 2 0 0.4 35 normal
particle splash ~ ~ ~-1.5 0 2 0 0.4 35 normal
particle splash ~ ~1.5 ~ 0 1 0 0.4 35 normal
particle splash ~ ~2.5 ~ 0 1 0 0.5 55 normal
particle splash ~ ~ ~-2.5 0 2 0 0.5 55 normal
particle splash ~-2.5 ~ ~ 0 2 0 0.5 55 normal
particle splash ~1.5 ~ ~ 0 2 0 0.5 55 normal

particle bubble ~0.0 ~1 ~-1.5 0 0.5 0 0.4 10 normal
particle bubble ~0.75 ~1 ~-1.299038 0 0.5 0 0.4 10 normal
particle bubble ~1.299038 ~1 ~-0.75 0 0.5 0 0.4 10 normal
particle bubble ~1.5 ~1 ~0.0 0 0.5 0 0.4 10 normal       
particle bubble ~0.75 ~1 ~1.299038 0 0.5 0 0.4 10 normal 
particle bubble ~-0.0 ~1 ~1.5 0 0.5 0 0.4 10 normal      
particle bubble ~-1.299038 ~1 ~0.75 0 0.5 0 0.4 10 normal
particle bubble ~-1.5 ~1 ~-0.0 0 0.5 0 0.4 10 normal
particle bubble ~-0.75 ~1 ~-1.299038 0 0.5 0 0.4 10 normal
summon minecraft:area_effect_cloud ~ ~0.4 ~ {WaitTime:1,Age:0,Duration:8,Color:554734,Radius:1,RadiusOnUse:1,RadiusPerTick:0.01}
# 匹配怪物
scoreboard players tag @e[r=5,c=4,tag=mob.asg,type=!player] add mob.target 
scoreboard players operation @e[r=5,c=4,tag=mob.target] uid = @s uid
scoreboard players set @e[r=5,c=4,tag=mob.asg] elemet.save.id 2
# 累加伤害
# 赋值伤害到怪物
scoreboard players set @e[r=5,c=4,tag=mob.target] hurt.water 20
scoreboard players set @e[r=5,c=4,tag=mob.target] hurt.armor 5
scoreboard players operation @e[r=5,c=4,tag=mob.target] hurt.water += @s elemet.water_lv
scoreboard players operation @e[r=5,c=4,tag=mob.target] hurt.water += @s more.water
scoreboard players operation @e[c=4,tag=elemet.skill] hurt.armor += @s damage.armor
# 怪物事件指向
execute @e[c=4,tag=mob.target] ~ ~ ~ function rpg:atk/elemet/water/1-3/3_mob
