# 分选称号
function sys:menu/carrot/name/select/承包鱼塘/main if @s[score_menu.trigger_min=6001,score_menu.trigger=6001,score_name.承包鱼塘_min=1]
# 提示
tellraw @s [{"text":"§r[§e称号系统§r]§7穿戴成功"}]