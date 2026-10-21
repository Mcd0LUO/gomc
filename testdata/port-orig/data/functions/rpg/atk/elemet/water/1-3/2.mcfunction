particle splash ~1.5 ~ ~ 0 2 0 0.4 40 normal
particle splash ~-1.5 ~ ~ 0 2 0 0.4 40 normal
particle splash ~ ~ ~1 0 2 0 0.4 40 normal
particle splash ~ ~ ~-1.5 0 2 0 0.4 40 normal
particle splash ~ ~1.5 ~ 0 1 0 0.4 40 normal

particle magicCrit ~0.0 ~0.1 ~-1.5 0 0.5 0 0.4 10 normal
particle magicCrit ~0.75 ~0.1 ~-1.299038 0 0.5 0 0.4 10 normal
particle magicCrit ~1.299038 ~0.1 ~0.75 0 0.5 0 0.4 10 normal
particle magicCrit ~-0.75 ~0.1 ~1.299038 0 0.5 0 0.4 10 normal
particle magicCrit ~-1.299038 ~0.1 ~0.75 0 0.5 0 0.4 10 normal
particle magicCrit ~-0.75 ~0.1 ~-1.299038 0 0.5 0 0.4 10 normal
summon minecraft:area_effect_cloud ~ ~0.4 ~ {WaitTime:1,Age:0,Duration:8,Color:554734,Radius:1,RadiusOnUse:1,RadiusPerTick:0.01}
particle bubble ~0.0 ~1 ~-1.5 0 0.5 0 0.4 10 normal
particle bubble ~0.75 ~1 ~-1.299038 0 0.5 0 0.4 10 normal
particle bubble ~1.299038 ~1 ~-0.75 0 0.5 0 0.4 10 normal
particle bubble ~1.5 ~1 ~0.0 0 0.5 0 0.4 10 normal       
particle bubble ~1.299038 ~1 ~0.75 0 0.5 0 0.4 10 normal 
particle bubble ~0.75 ~1 ~1.299038 0 0.5 0 0.4 10 normal 
particle bubble ~-0.0 ~1 ~1.5 0 0.5 0 0.4 10 normal      
particle bubble ~-1.299038 ~1 ~0.75 0 0.5 0 0.4 10 normal
particle bubble ~-1.5 ~1 ~-0.0 0 0.5 0 0.4 10 normal
particle bubble ~-1.299038 ~1 ~-0.75 0 0.5 0 0.4 10 normal
particle bubble ~-0.75 ~1 ~-1.299038 0 0.5 0 0.4 10 normal

# 匹配怪物
scoreboard players tag @e[r=3,c=4,tag=mob.asg,type=!player] add mob.target 
scoreboard players operation @e[r=3,c=4,tag=mob.target] uid = @s uid
# 累加伤害
scoreboard players set @s damage.water 16
scoreboard players operation @s damage.water += @s elemet.water_lv
scoreboard players operation @s damage.water += @s more.water
# 赋值伤害到怪物
scoreboard players operation @e[r=3,c=4,tag=mob.target] hurt.water = @s damage.water
# 怪物事件指向
execute @e[c=4,tag=mob.target] ~ ~ ~ function rpg:atk/elemet/water/1-3/2_mob
