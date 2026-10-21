# 赋值
scoreboard players operation @s damage.real += @s pet.lv
scoreboard players operation @e[tag=mob.asg,c=1] uid = @s pet.uid
scoreboard players operation @e[tag=mob.asg,c=1] hurt.real = @s damage.real
# 怪物受伤运算
execute @e[tag=mob.asg,c=1] ~ ~ ~ function rpg:monster/hurt/other/pet
# 反馈
particle iconcrack ~ ~1 ~ 0.1 0.2 0.1 0.08 20 normal @a[r=15] 152s