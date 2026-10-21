# 非第一次登录的玩家每次登录游戏都会调用此接口
# 用于重载玩家的各项数据

# 设置标题时间
title @s times 0 8 0
# 获取自身版本值
scoreboard players operation @s arg1 = @s version
scoreboard players operation @s arg1 -= $VERDSION$ version
function sys:basic/reload/versionupdate if @s[score_temp=-1]
scoreboard players reset @s arg1
# 重载称号
scoreboard players reset @s name.id
scoreboard players reset @s menu.trigger
function sys:menu/carrot/name/name_main
# 重置数据
scoreboard players reset @s leaveGame
tellraw @s ["§r[§e系统§r]:§7重新登录,§r数据重载成功"]