playsound entity.zombie.attack_iron_door master @s ~0 ~0 ~0 3 0.2 0
playsound entity.wither_skeleton.ambient master @s ~0 ~0 ~0 3 0.1 0
playsound entity.wither_skeleton.step master @s ~0 ~0 ~0 3 1 0
playsound entity.wither_skeleton.death master @s ~0 ~0 ~0 3 1 0
playsound entity.zombie_villager.cure master @s ~0 ~0 ~0 3 1 0
# 匹配怪物
scoreboard players tag @e[r=12,c=1,tag=mob.asg,type=!player] add mob.target 
scoreboard players operation @e[c=1,tag=mob.target] uid = @s uid
scoreboard players set @e[c=1,tag=mob.target] elemet.save.id 4
# 累加伤害
scoreboard players set @s damage.thunder 27
scoreboard players operation @s damage.thunder += @s elemet.thund_lv
# 赋值伤害到怪物
scoreboard players operation @e[c=1,tag=mob.target] hurt.thunder = @s damage.thunder
scoreboard players set @e[r=5,c=4,tag=mob.target] hurt.armor 5
# 怪物事件指向
execute @e[c=1,tag=mob.target] ~ ~ ~ function rpg:atk/elemet/thunder/1-3/3_mob