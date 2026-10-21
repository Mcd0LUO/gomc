scoreboard players tag @s add skill.火 {SelectedItem:{tag:{Tags:["tr","fire"]}}}
function sys:skill/select/fire/main if @s[tag=skill.火]
scoreboard players tag @s add skill.水 {SelectedItem:{tag:{Tags:["tr","water"]}}}
function sys:skill/select/water/main if @s[tag=skill.水]
scoreboard players tag @s add skill.冰 {SelectedItem:{tag:{Tags:["tr","ice"]}}}
function sys:skill/select/ice/main if @s[tag=skill.冰]
scoreboard players tag @s add skill.雷 {SelectedItem:{tag:{Tags:["tr","thunder"]}}}
function sys:skill/select/thunder/main if @s[tag=skill.雷]

scoreboard players reset @s menu.trigger
scoreboard players set @s menu.use 0
scoreboard players set @s weapon.use 0