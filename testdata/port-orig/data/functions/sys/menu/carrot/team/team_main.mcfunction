# 队伍系统主类


# 需无队伍类
function sys:menu/carrot/team/no_team if @s[tag=!队伍中,score_menu.trigger_min=3001,score_menu.trigger=3999]
# 需已有队伍类
function sys:menu/carrot/team/have_team if @s[tag=队伍中,score_menu.trigger_min=3001,score_menu.trigger=3999]