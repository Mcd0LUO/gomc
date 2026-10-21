# sort-max
# 赋值所有玩家的 arg1 = 待sort分数

# sort算法 
function sys:math/sort/max/max
# 记录排名
scoreboard players add $sort$ sort 1
scoreboard players operation @a[tag=最高分数] sort = $sort$ sort
# 输出排名
function sys:math/sort/max/print
# 移除参与
scoreboard players tag @a[tag=最高分数] remove 最高分数
# 抛出异常
function sys:math/sort/max/trycrash unless @a[tag=已参与]
# 递归
function sys:math/sort/max/run if @a[tag=!已参与]
function sys:math/sort/max/end unless @a[tag=!已参与]