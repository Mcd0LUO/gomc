# 获取攻击怪物的玩家api
scoreboard players operation @a[tag=self.atk,r=12] temp.uid_match = @s uid
execute @a[tag=self.atk,r=12] ~ ~ ~ scoreboard players operation @s temp.uid_match -= @s uid
# [score_temp.uidmatch_min=0,score_temp.uidmatch=0]