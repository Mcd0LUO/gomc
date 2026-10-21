# 获取接受到的元素强效id
scoreboard players operation @s elemet.save.id = @e[c=1,tag=mob.asg,dy=-1] elemet.save.id
# 重运算超强效元素挂留
scoreboard players set @s[tag=elemet.strong] elemet.saveTotal 1 
# 初始化更新元素显示位置
function rpg:monster/hurt/elemet/addElemet unless @s[score_elemet.saveTotal_min=2]
function rpg:monster/hurt/elemet/fire/add_fire if @s[score_elemet.save.id_min=1,score_elemet.save.id=1,score_elemet.saveTotal=1]
# 元素聚变
function rpg:monster/hurt/elemet/fire/fission/water if @s[tag=elemet.saveWater]
function rpg:monster/hurt/elemet/fire/fission/ice if @s[tag=elemet.saveIce]
function rpg:monster/hurt/elemet/fire/fission/thunder if @s[tag=elemet.saveThunder]

#tellraw @a [{"score":{"objective":"elemet.save.id","name":"@s"},"color":"red"}]
# 重置
scoreboard players reset @s elemet.save.id