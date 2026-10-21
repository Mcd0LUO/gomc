scoreboard players operation @s health = @s health.max
scoreboard players operation @s mp = @s mp.max
scoreboard teams leave @s
function rpg:hurt/player/death/fail_boss if @s[tag=团本中]

function sys:math/random/1-200
execute @s[score_return_min=1,score_return=20] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"§7在与",{"selector":"@e[tag=attribute.name,c=1]"},"§7的战斗中不幸飞升"]
execute @s[score_return_min=21,score_return=40] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"§7在与",{"selector":"@e[tag=attribute.name,c=1]"},"§7的战斗时突然想起自己不会用魔法"]
execute @s[score_return_min=41,score_return=60] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"§7掉落的火之祈福让",{"selector":"@e[tag=attribute.name,c=1]"},"§7感受到了火神的认可"]
execute @s[score_return_min=61,score_return=80] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"被",{"selector":"@e[tag=attribute.name,c=1]"},"§7轰到了提瓦特"]
execute @s[score_return_min=81,score_return=100] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"与",{"selector":"@e[tag=attribute.name,c=1]"},"§7不共戴天"]
execute @s[score_return_min=101,score_return=120] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"§7在玩金铲铲时发现自己被",{"selector":"@e[tag=attribute.name,c=1]"},"§7撕碎"]
execute @s[score_return_min=121,score_return=140] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"§7试图与",{"selector":"@e[tag=attribute.name,c=1]"},"§7贴贴时被发现"]
execute @s[score_return_min=141,score_return=160] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@s"},"§7的元素对",{"selector":"@e[tag=attribute.name,c=1]"},"§7不痛不痒"]
execute @s[score_return_min=161,score_return=180] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@e[tag=attribute.name,c=1]"},"§7在击杀",{"selector":"@s"},"§7后宣布将主宰世界"]
execute @s[score_return_min=181,score_return=200] ~ ~ ~ tellraw @a ["§c⚔",{"selector":"@e[tag=attribute.name,c=1]"},"§7使用了",{"selector":"@s"},"§7消失术,效果拔群!"]
# 移除debuff
scoreboard players reset @s debuff.armor
# 死亡
kill @s