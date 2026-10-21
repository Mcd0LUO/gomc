# 怪物受水元素飞行物技能接口
# 匹配玩家
scoreboard players operation @a[r=12] temp.uid.match = @s uid
execute @a[r=12] ~ ~ ~ scoreboard players operation @s temp.uid.match -= @s uid
# 赋值
scoreboard players operation @e[c=1,tag=mob.asg] hurt.fire = @s damage.fire
scoreboard players operation @e[c=1,tag=mob.asg] hurt.armor = @s damage.armor
scoreboard players operation @e[c=1,tag=mob.asg] elemet.save.id = @s elemet.save.id
scoreboard players operation @e[c=1,tag=mob.asg] hurt.real = @s damage.real
# 运算
execute @e[c=1,tag=mob.asg] ~ ~ ~ function rpg:monster/hurt/elemet/fire/main_fire
# reset
execute @a[r=12] ~ ~ ~ scoreboard players reset @s hurt.temp
kill @s