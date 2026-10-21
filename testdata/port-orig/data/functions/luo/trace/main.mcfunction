## 法球追踪主类
    #@param arg1: 追踪次数
    scoreboard players tag @s add api.trace_main
    execute @s ~ ~ ~ function luo:trace/Exception unless @e[r=12,tag=api.trace_target]
    execute @s ~ ~ ~ function luo:trace/cyc if @e[r=12,tag=api.trace_target]
    execute @s ~ ~ ~ function luo:trace/end if @e[r=1,tag=api.trace_target]
#