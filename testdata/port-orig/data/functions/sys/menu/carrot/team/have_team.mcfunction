# 有队伍玩家接口

# 抛异
function sys:menu/carrot/team/crash_have_team if @s[score_menu.trigger_min=3001,score_menu.trigger=3002]
# 队长玩家点击离开队伍
function sys:menu/carrot/team/crash_have_admin if @s[score_menu.trigger_min=3003,score_menu.trigger=3003,tag=队长]
# 非队长点击管理队伍/邀请玩家
function sys:menu/carrot/team/crash_not_admin if @s[score_menu.trigger_min=3004,score_menu.trigger=3005,tag=!队长]
# 离开队伍
function sys:menu/carrot/team/leave/main if @s[score_menu.trigger_min=3003,score_menu.trigger=3003,tag=!队长]
# 管理队伍
function sys:menu/carrot/team/manage/main if @s[score_menu.trigger_min=3004,score_menu.trigger=3004,tag=队长]
# 解散队伍
function sys:menu/carrot/team/manage/destory if @s[score_menu.trigger_min=3007,score_menu.trigger=3007,tag=队长]
# 踢出玩家
function sys:menu/carrot/team/manage/kick if @s[score_menu.trigger_min=3008,score_menu.trigger=3008,tag=队长]
