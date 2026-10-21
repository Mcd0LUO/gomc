# 抛异
scoreboard players set @s[tag=释放宠物] menu.trigger 0
tellraw @s[tag=释放宠物] ["§c✘ §eERROR:§7禁止重复召唤，只能同时存在§e一只§7活动的战兽"]
# sort
function sys:pet/sum/银月啸嗷 if @s[score_menu.trigger_min=5101,score_menu.trigger=5101,score_pet.now_min=1,score_pet.now=1]