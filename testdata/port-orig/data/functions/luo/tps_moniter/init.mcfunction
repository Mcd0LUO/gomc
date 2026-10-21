scoreboard objectives add TpsMoniter:TPS dummy 边界存储
scoreboard objectives add TpsMoniter:Temp dummy tps检测临时变量

kill @e[tag=tps_moniter.watcher]
summon armor_stand ~ ~1 ~ {Tags:["tps_moniter.watcher"],PersistenceRequired:1b,Marker:1b}
stats entity @e[tag=tps_moniter.watcher] set QueryResult @e[tag=tps_moniter.watcher] TpsMoniter:TPS
scoreboard players set @e[tag=tps_moniter.watcher] TpsMoniter:TPS 0
