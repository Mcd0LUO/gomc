# 玩家受击接口

# ini
scoreboard players set @s math.operation1 0
# fire
scoreboard players operation @s math.operation1 = @e[c=1,tag=mob.asg] damage.fire
scoreboard players operation @s math.operation1 -= @s armor.fire
scoreboard players operation @s[score_math.operation1_min=1] hurt.total += @s math.operation1
# water
scoreboard players set @s math.operation1 0
scoreboard players operation @s math.operation1 = @e[c=1,tag=mob.asg] damage.water
scoreboard players operation @s math.operation1 -= @s armor.water
scoreboard players operation @s[score_math.operation1_min=1] hurt.total += @s math.operation1
#ice
scoreboard players set @s math.operation1 0
scoreboard players operation @s math.operation1 = @e[c=1,tag=mob.asg] damage.ice
scoreboard players operation @s math.operation1 -= @s armor.ice
scoreboard players operation @s[score_math.operation1_min=1] hurt.total += @s math.operation1
# thunder
scoreboard players set @s math.operation1 0
scoreboard players operation @s math.operation1 = @e[c=1,tag=mob.asg] damage.thunder
scoreboard players operation @s math.operation1 -= @s armor.thunder
scoreboard players operation @s[score_math.operation1_min=1] hurt.total += @s math.operation1
# 全魔抗运算
scoreboard players operation @s arg1 = @s hurt.total
scoreboard players operation @s arg1 *= armor.player.all
scoreboard players operation @s arg1 /= 100 num
# debuff增伤运算
scoreboard players operation @s arg2 = @s hurt.total
scoreboard players operation @s arg2 *= @s debuff.armor
scoreboard players operation @s arg2 /= 100 num
scoreboard players operation @s[score_arg2_min=1] hurt.total += @s arg2
# 扣除hp
scoreboard players operation @s hurt.total -= @s arg1
scoreboard players operation @s[score_hurt.total_min=1] health -= @s hurt.total
# 玩家死亡事件 
function rpg:hurt/player/death/main if @s[score_health=0]
# reset
scoreboard players reset @s event
scoreboard players reset @s hurt.total
advancement revoke @s from main:adv/rpg/main
scoreboard players reset @s math.operation1