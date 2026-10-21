# 踢出不安全的玩家以防止服务器崩溃
    scoreboard players operation #temp temp1 = @TPS_TEST TpsMoniter:TPS
    scoreboard players operation #temp temp2 = @TPS_TEST TpsMoniter:TPS
    scoreboard players operation #temp temp1 /= math.100 math
    scoreboard players operation #temp temp2 %= math.100 math
    tellraw @a ["§c❖ 服务器当前TPS ",{"score":{"name":"#temp", "objective":"temp1"}},"§a.",{"score":{"name":"#temp", "objective":"temp2"}}," §c低于正常值,疑似遭到攻击,正在尝试清理异常玩家..."]

    function luo:tps_moniter/heuristic/main

#