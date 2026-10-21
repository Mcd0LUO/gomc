# 获取接受到的元素强效id
scoreboard players operation @s elemet.save.id = @e[c=1,tag=mob.asg,dy=-2] elemet.save.id
# 重运算超强效元素挂留
scoreboard players set @s[tag=elemet.strong] elemet.saveTotal 1 
# 初始化更新元素显示位置
function rpg:monster/hurt/elemet/addElemet unless @s[score_elemet.saveTotal_min=2]
function rpg:monster/hurt/elemet/water/add_water if @s[score_elemet.save.id_min=2,score_elemet.save.id=2,score_elemet.saveTotal=1]
# 元素聚变
function rpg:monster/hurt/elemet/water/fission/fire if @s[tag=elemet.saveFire]
function rpg:monster/hurt/elemet/water/fission/ice if @s[tag=elemet.saveIce]
function rpg:monster/hurt/elemet/water/fission/thunder if @s[tag=elemet.saveThunder]

#tellraw @a [{"score":{"objective":"elemet.save.id","name":"@s"},"color":"red"}]
# 重置
scoreboard players reset @s elemet.save.id