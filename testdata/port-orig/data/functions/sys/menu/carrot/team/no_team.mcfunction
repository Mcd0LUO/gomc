# 无队伍玩家运行

# 抛异
# 点击离开队伍但无队伍 # 无队伍玩家点击管理/邀请玩家
function sys:menu/carrot/team/crash_no_team if @s[score_menu.trigger_min=3003,score_menu.trigger=3005]
# 创建队伍
function sys:menu/carrot/team/create/main if @s[score_menu.trigger_min=3001,score_menu.trigger=3001]
# 加入队伍
function sys:menu/carrot/team/join/main if @s[score_menu.trigger_min=3002,score_menu.trigger=3002]