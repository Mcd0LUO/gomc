## 测算tps

    scoreboard players set @TPS_TEST TpsMoniter:TPS 40000000
    execute @e[type=armor_stand,c=1,tag=tps_moniter.watcher] ~ ~ ~ worldborder get
    scoreboard players remove @e[type=armor_stand,c=1,tag=tps_moniter.watcher] TpsMoniter:TPS 40000000
    scoreboard players operation @TPS_TEST TpsMoniter:TPS /= @e[type=armor_stand,c=1,tag=tps_moniter.watcher] TpsMoniter:TPS

    worldborder set 40000000
    worldborder set 60000000 1000

    ## 反崩看门狗

    scoreboard players operation @e[tag=tps_moniter.watcher] TpsMoniter:Temp = @TPS_TEST TpsMoniter:TPS
    function luo:tps_moniter/watcher if @e[tag=tps_moniter.watcher,score_TpsMoniter:Temp=1400]

    tellraw @a [{"score":{"name":"", "objective":"objective"}}]
   
#