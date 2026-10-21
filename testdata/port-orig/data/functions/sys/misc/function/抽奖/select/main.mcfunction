
# 退出
function sys:misc/function/抽奖/select/exit if @s[rxm=-56,rx=-41,rym=167,ry=192]
# 确定
function sys:misc/function/抽奖/select/confirm if @s[rxm=34,rx=49,rym=165,ry=197]
# 单抽
function sys:misc/function/抽奖/select/单抽 if @s[rxm=32,rx=43,rym=147,ry=160]
# 十连
function sys:misc/function/抽奖/select/十连 if @s[rxm=32,rx=43,rym=-153,ry=-141]
# 选择左
function sys:misc/function/抽奖/select/切换_左 if @s[rxm=-17,rx=19,rym=124,ry=148]
# 选择右
function sys:misc/function/抽奖/select/切换_右 if @s[rxm=-17,rx=19,rym=-144,ry=-122]

scoreboard players reset @s talk