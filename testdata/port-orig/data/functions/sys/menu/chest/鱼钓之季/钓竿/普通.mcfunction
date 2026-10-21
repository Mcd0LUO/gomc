scoreboard players operation @s arg1 = @s money
scoreboard players remove @a arg1 300
#
function sys:misc/give/活动/鱼钓之季/普通钓竿 if @s[score_arg1_min=300]
tellraw @s[score_arg1_min=300] ["§r[§e鱼钓之季§r]:§7成功购买§e普通§7钓竿,钓鱼快乐啊~"]
tellraw @s[score_arg1=299] ["§r[§e鱼钓之季§r]:无效的操作，条件不符合"]
scoreboard players remove @s[score_arg1_min=300] money 300

function sys:api/粒子/灵子崩坏/+y if @s[score_arg1_min=300]
#
scoreboard players reset @s arg1