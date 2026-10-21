scoreboard players set @s arg1 1
scoreboard players set @s arg2 420
function sys:math/random/run
execute @s[score_return_min=1,score_return=100] ~ ~ ~ function sys:misc/give/活动/鱼钓之季/小龙鱼
execute @s[score_return_min=101,score_return=180] ~ ~ ~ function sys:misc/give/活动/鱼钓之季/小龙糕
execute @s[score_return_min=201,score_return=240] ~ ~ ~ function sys:misc/give/活动/鱼钓之季/龙年大吉
execute @s[score_return_min=241,score_return=260] ~ ~ ~ function sys:misc/give/活动/鱼钓之季/大师钓竿
execute @s[score_return_min=271,score_return=400] ~ ~ ~ function sys:misc/give/活动/鱼钓之季/饺子
execute @s[score_return_min=41,score_return=420] ~ ~ ~ function sys:misc/give/活动/鱼钓之季/小龙鱼
clear @s bread 0 128 {Tags:["饺子"]}