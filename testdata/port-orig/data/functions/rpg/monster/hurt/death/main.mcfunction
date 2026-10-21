# 增加击杀点
scoreboard players add @a[score_temp.uid.match_min=0,score_temp.uid.match=0] data.kill 1
tellraw @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ["§e⚔§7击杀怪物，§e击杀点§a+1，§7当前拥有击杀点§e҉",{"score":{"objective":"data.kill","name": "@a[score_temp.uid.match_min=0,score_temp.uid.match=0]"},"color": "red"}]

kill @s
kill @e[c=1,type=sheep,dy=4]