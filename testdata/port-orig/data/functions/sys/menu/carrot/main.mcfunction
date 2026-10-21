tellraw @a[score_menu.use_min=1] [{"text":"┍───────────────菜单─────────────┑\n│ "},{"text":"┅§e返回主城┅","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 101"},"hoverEvent":{"action":"show_text","value":"§7紫砂回城"}},{"text":"  ┅§b属性表单§r┅","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 102"},"hoverEvent":{"action":"show_text","value":"§7生成属性报告表"}},{"text":"  ┅§a队伍系统§a┅","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 103"},"hoverEvent":{"action":"show_text","value":"§7管理/加入队伍"}}," │\n│────────────────────────────────│\n│ ",{"text":"┅§c传送菜单§r┅","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 104"},"hoverEvent":{"action":"show_text","value":"§7超位_传送魔法"}},{"text":"  ┅§d战兽系统§r┅","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 105"},"hoverEvent":{"action":"show_text","value":"§r元素契约兽管理"}},{"text":"  ┅§3称号系统§r┅ │","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 106"},"hoverEvent":{"action":"show_text","value":"§r...\n"}},"┕───────────────菜单─────────────┙"]

# 菜单第一页选中
function sys:menu/carrot/page_1 if @a[score_menu.trigger_min=101,score_menu.trigger=199]
# 队伍系统
execute @a[score_menu.trigger_min=3001,score_menu.trigger=3999] ~ ~ ~ function sys:menu/carrot/team/team_main
# 传送系统
function sys:menu/carrot/tp/tp_main if @a[score_menu.trigger_min=4001,score_menu.trigger=4099]
# 团本
function sys:menu/carrot/tp/tp_main if @a[score_menu.trigger_min=4001,score_menu.trigger=4099]
# 称号系统
execute @a[score_menu.trigger_min=6001,score_menu.trigger=6999] ~ ~ ~ function sys:menu/carrot/name/name_main
# 宠物系统
execute @a[score_menu.trigger_min=5001,score_menu.trigger=5999] ~ ~ ~ function sys:menu/carrot/pet/main
# reset
scoreboard players reset @a[score_menu.use_min=1,score_menu.trigger_min=1] weapon.use
scoreboard players reset @a[score_menu.use_min=1,score_menu.trigger_min=1] menu.use
scoreboard players reset @a[score_menu.trigger_min=1] menu.trigger

scoreboard players enable @a menu.trigger