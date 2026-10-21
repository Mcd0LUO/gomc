# 分配零时uid
scoreboard players add #team.list.key# team.key 1
scoreboard players add @a[tag=队长] team.key 0
scoreboard players tag @r[type=player,tag=队长,score_team.key_min=0,score_team.key=0,c=1] add team.key.match
scoreboard players operation @a[tag=team.key.match,score_team.key_min=0,score_team.key=0] team.key = #team.list.key# team.key
# 递归
function sys:menu/carrot/team/join/asg_team if @a[tag=队长,score_team.key_min=0,score_team.key=0]
# 结束递归
function sys:menu/carrot/team/join/asg_end unless @a[tag=队长,score_team.key_min=0,score_team.key=0]