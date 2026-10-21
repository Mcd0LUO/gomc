scoreboard players set @s arg1 1
scoreboard players set @s arg2 100
function sys:math/random/run
scoreboard players add @s[score_return_min=1,score_return=40] elemet.thund_lv 1
tellraw @s[score_return_min=1,score_return=40] ["§r[§e歌莉娅§r]:§7感受到了§6雷神§7的认可，祈福§b成功!雷元素等级提升1级 现在为",{"score":{"objective":"elemet.thund_lv","name":"@s"},"color":"red"}]
playsound block.anvil.use master @s[score_return_min=1,score_return=40] ~0 ~0 ~0 1 1 0
tellraw @s[score_return_min=41,score_return=100] ["§r[§e歌莉娅§r]:§7抱歉,§6雷神§7没有任何反应，祈福§c失败 :("]
particle totem ~ ~1 ~ 0 0 0 0.25 40 normal
playsound block.anvil.destroy master @s[score_return_min=41,score_return=100] ~0 ~0 ~0 1 1 0
scoreboard players reset @s return
scoreboard players set @s stats.item -1