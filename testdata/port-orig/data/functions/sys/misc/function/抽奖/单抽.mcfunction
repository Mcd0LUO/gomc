scoreboard players remove @s dq 180

scoreboard players set @s arg1 1
scoreboard players set @s arg2 10000
function sys:math/random/run

function sys:misc/function/抽奖/pool/普通 if @s[score_return_min=1,score_return=6000]
function sys:misc/function/抽奖/pool/中级 if @s[score_return_min=6001,score_return=8800]
function sys:misc/function/抽奖/pool/高级 if @s[score_return_min=8801,score_return=9800]
function sys:misc/function/抽奖/pool/大奖 if @s[score_return_min=9801]

scoreboard players reset @s return