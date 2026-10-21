clear @s nether_star 0 0 {Tags:["潜蛟腾跃","栏位:1","淬炼:0"]}
scoreboard players tag @s[score_stats.item_min=2] add 允许淬炼 {SelectedItem:{tag:{Tags:["潜蛟腾跃","栏位:1","淬炼:0"]}}}
function sys:menu/chest/淬炼/潜蛟腾跃/pass if @s[tag=允许淬炼]

function sys:menu/chest/淬炼/update
tellraw @s[tag=!允许淬炼] ["§r[§e洛必达§r]:§7条件不符，无法淬炼"]
scoreboard players tag @s remove 允许淬炼