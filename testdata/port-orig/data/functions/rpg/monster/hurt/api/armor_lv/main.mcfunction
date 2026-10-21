# 根据传入的百分比修改自身的队伍
# 一般面向怪物
function rpg:monster/hurt/api/armor_lv/1-50 if @s[score_return_min=1,score_return=50]
function rpg:monster/hurt/api/armor_lv/51-100 if @s[score_return_min=51,score_return=100]