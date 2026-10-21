# atk of all elemet

# __ini
scoreboard players add @s weapon.useTotal 1
scoreboard players add @s angry 10
scoreboard players operation @s angry < 150 num
scoreboard players tag @s add self.atk
# select
scoreboard players set @s weapon.eleid 1 {Inventory:[{tag:{Tags:["wp_tr","fire"]},Slot:-106b}]}
scoreboard players set @s weapon.eleid 2 {Inventory:[{tag:{Tags:["wp_tr","water"]},Slot:-106b}]}
scoreboard players set @s weapon.eleid 3 {Inventory:[{tag:{Tags:["wp_tr","ice"]},Slot:-106b}]}
scoreboard players set @s weapon.eleid 4 {Inventory:[{tag:{Tags:["wp_tr","thunder"]},Slot:-106b}]}
function rpg:atk/elemet/fire/main if @s[score_weapon.eleid_min=1,score_weapon.eleid=1]
function rpg:atk/elemet/water/main if @s[score_weapon.eleid_min=2,score_weapon.eleid=2]
function rpg:atk/elemet/ice/main if @s[score_weapon.eleid_min=3,score_weapon.eleid=3]
function rpg:atk/elemet/thunder/main if @s[score_weapon.eleid_min=4,score_weapon.eleid=4]
# reset/add
scoreboard players reset @s weapon.use
scoreboard players reset @s menu.use
scoreboard players reset @s weapon.eleid
scoreboard players reset @s[score_weapon.useTotal_min=3] weapon.useTotal
scoreboard players tag @s remove self.atk