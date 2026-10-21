# sort-min
# 赋值所有玩家的 arg1 = 待sort分数

# sort算法 
function sys:math/sort/min/min
# 记录排名
scoreboard players add $sort$ sort 1
scoreboard players operation @a[tag=最低分数] sort = $sort$ sort
# 输出排名
function sys:math/sort/min/print
# 移除参与
scoreboard players tag @a[tag=最低分数] remove 最低分数
# 抛出异常
function sys:math/sort/min/trycrash unless @a[tag=已参与]
# 递归
function sys:math/sort/min/run if @a[tag=!已参与]
function sys:math/sort/min/end unless @a[tag=!已参与]