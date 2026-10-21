# 清除数显 文显盔甲架方法
# 在main命名空间下的主函数调用即可
scoreboard players add @e[tag=数显] temp 1
kill @e[score_temp_min=2,tag=数显]