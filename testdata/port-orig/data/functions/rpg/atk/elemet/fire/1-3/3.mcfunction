# 玩家反馈
playsound entity.spider.step master @s ~0 ~0 ~0 1 1 0
playsound entity.blaze.shoot master @s ~0 ~0 ~0 1 1 0
playsound entity.creeper.primed master @s ~0 ~0 ~0 1 1 0
playsound entity.wither.break_block master @s ~0 ~0 ~0 1 1 0

summon armor_stand ~ ~ ~ {Small:1b,Marker:1b,Invisible:1b,Tags:["elemet.skill","fire","fire.1"]}
summon armor_stand ~ ~ ~ {Small:1b,Marker:1b,Invisible:1b,Tags:["elemet.skill","fire","fire.2"]}
scoreboard players operation @e[c=1,tag=elemet.skill] uid = @s uid
# 运算伤害
scoreboard players set @e[c=2,tag=elemet.skill] damage.fire 25
scoreboard players operation @e[c=2,tag=elemet.skill] damage.fire += @s elemet.fire_lv
scoreboard players operation @e[c=2,tag=elemet.skill] damage.fire += @s more.fire
scoreboard players operation @e[c=2,tag=elemet.skill] damage.armor += @s damage.armor
#
scoreboard players set @e[c=2,tag=elemet.skill] elemet.save.id 1
tp @e[c=2,tag=elemet.skill] @s
tp @e[c=2,tag=elemet.skill] ~ ~2 ~
execute @e[c=1,tag=fire.1] ~ ~ ~ function sys:api/视角选择器/horizontal/tp/1/main if @s[tag=elemet.skill]
execute @e[c=1,tag=fire.2] ~ ~ ~ function sys:api/视角选择器/horizontal/motion/3/main if @s[tag=elemet.skill]

effect @s minecraft:slowness 1 1 true