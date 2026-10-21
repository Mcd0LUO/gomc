# 玩家反馈
playsound entity.firework.twinkle master @s ~0 ~0 ~0 1 1 0
playsound block.fire.extinguish master @s ~0 ~0 ~0 1 1 0
playsound entity.skeleton.hurt master @s ~0 ~0 ~0 1 1 0
# 匹配怪物
scoreboard players tag @e[r=12,c=2,tag=mob.asg,type=!player] add mob.target 
scoreboard players operation @e[c=2,tag=mob.target] uid = @s uid
# 累加伤害
scoreboard players set @s damage.fire 18
scoreboard players operation @s damage.fire += @s elemet.fire_lv
scoreboard players operation @s damage.fire += @s more.fire
# 赋值伤害到怪物
scoreboard players operation @e[c=2,tag=mob.target] hurt.fire = @s damage.fire
# 怪物事件指向
execute @e[c=3,tag=mob.target] ~ ~ ~ function rpg:atk/elemet/fire/1-3/2_mob