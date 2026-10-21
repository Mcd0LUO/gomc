# water_Atk_main
scoreboard players reset @s[score_mp.now=4] weapon.useTotal
tellraw @s[score_mp.now=4] ["§r[§e系统§r]:§7蓝量不足，请等待回蓝"]
scoreboard players remove @s[score_mp.now_min=5] mp.now 5
# 1-4 atk
function rpg:atk/elemet/water/1-3/1 if @s[score_weapon.useTotal_min=1,score_weapon.useTotal=1]
function rpg:atk/elemet/water/1-3/2 if @s[score_weapon.useTotal_min=2,score_weapon.useTotal=2]
function rpg:atk/elemet/water/1-3/3 if @s[score_weapon.useTotal_min=3,score_weapon.useTotal=3]
function rpg:atk/elemet/water/1-3/angry if @s[score_angry_min=120,score_weapon.useTotal=120]
# 
playsound entity.hostile.splash master @s ~0 ~0 ~0 1 1 0
playsound entity.hostile.swim master @s ~0 ~0 ~0 1 2 0
# 
scoreboard players reset @s weapon.use
scoreboard players reset @s menu.use