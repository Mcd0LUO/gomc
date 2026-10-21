## 选取不重复数字 0-8
# 需要手动调用多次
scoreboard teams join random.sampled @e[type=armor_stand,tag=rand.9]
scoreboard players operation @s return = @r[type=armor_stand,tag=range.0_9,team=!random.sampled] math
scoreboard players operation @e[type=armor_stand,tag=range.0_9] temp1 = @s return 
execute @e[type=armor_stand,tag=range.0_9] ~ ~ ~ scoreboard players operation @s temp1 -= @s math
scoreboard teams join random.sampled @e[type=armor_stand,tag=range.0_9,score_temp1_min=0,score_temp1=0]