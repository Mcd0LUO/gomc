scoreboard players reset @s elemet.saveTotal
entitydata @s {CustomName:""}
execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s arg1 *= 2 num

execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s hurt.armor * @s 盾伤倍率

execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s armor.all.now -= @s hurt.armor

particle largeexplode ~ ~-1 ~ 0 0 0 0.25 1 normal
particle fireworksSpark ~ ~ ~ 0 0 0 0.25 10 normal

particle sweepAttack ~0.0 ~ ~-2.0 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.147153 ~ ~-1.638304 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.879385 ~ ~-0.68404 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.931852 ~ ~0.517638 -0.0 0 -0.0 0.3 1 normal

particle sweepAttack ~-1.414214 ~ ~-1.414214 0.0 0 0.0 0.3 1 normal
particle sweepAttack ~-0.347296 ~ ~-1.969616 0.0 0 0.0 0.3 1 normal
summon arrow ~ ~2.5 ~ {Motion:[0.0,-3.0,0.0]}

execute @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ~ ~ ~ function rpg:monster/hurt/normal/thunder/backToPlayer

tellraw @a[score_temp.uid.match_min=0,score_temp.uid.match=0] [["§7[§c§l༗ §r| §e§l⚡§7]->超载 剧烈爆炸 元素溅射"]]
scoreboard players tag @s remove elemet.saveThunder

summon lightning_bolt ~ ~4.3 ~
particle flame ~-0.75 ~0 ~0.0 0 0 0 0 1 normal
particle flame ~-0.9050999681267564 ~0 ~4.085007963455079 0 0 0 0 1 normal
particle flame ~1.3311049049897925 ~0 ~-0.9716043466392063 0 0 0 0 1 normal
particle flame ~2.20728750021509 ~0 ~-2.932802416010863 0 0 0 0 1 normal
particle flame ~-2.6550507757909885 ~0 ~1.0328381860936036 0 0 0 0 1 normal
particle flame ~-2.334963470471356 ~0 ~1.39296299788045 0 0 0 0 1 normal
particle flame ~3.7545700824935535 ~0 ~0.0912936066036304 0 0 0 0 1 normal
particle flame ~1.4281324518536902 ~0 ~-0.465676527463152 0 0 0 0 1 normal
particle flame ~-3.7855389873971537 ~0 ~-1.8452630991649515 0 0 0 0 1 normal
particle flame ~-0.33174741884758974 ~0 ~0.6952333845286192 0 0 0 0 1 normal
particle flame ~2.6133633903204894 ~0 ~3.2233737891661405 0 0 0 0 1 normal
particle flame ~0.006061197566168852 ~0 ~-1.7949531471810145 0 0 0 0 1 normal
particle flame ~-0.9084785077704557 ~0 ~-3.462230250440782 0 0 0 0 1 normal
particle flame ~-0.8574286104597395 ~0 ~2.8483269626383074 0 0 0 0 1 normal
particle flame ~-0.31523711548807065 ~0 ~2.565724225878911 0 0 0 0 1 normal
particle flame ~2.453201637537061 ~0 ~-2.9469916233137816 0 0 0 0 1 normal
particle flame ~0.39944336254390045 ~0 ~-1.298949212843601 0 0 0 0 1 normal
particle flame ~-3.818285761032864 ~0 ~1.8235405857654625 0 0 0 0 1 normal
particle flame ~0.5064105649866323 ~0 ~0.6553268274858048 0 0 0 0 1 normal
particle flame ~4.10799206638441 ~0 ~-0.04074408772556681 0 0 0 0 1 normal
particle flame ~-1.5527702326104182 ~0 ~-1.166131029393294 0 0 0 0 1 normal
particle flame ~-3.191144070122427 ~0 ~-1.393976569622992 0 0 0 0 1 normal
particle flame ~1.7829176581911084 ~0 ~2.5307580192026418 0 0 0 0 1 normal
particle flame ~1.7395541650734936 ~0 ~1.7218483036477275 0 0 0 0 1 normal
particle flame ~-0.8003433700025877 ~0 ~-3.823892790924206 0 0 0 0 1 normal
particle flame ~-0.7619160521640718 ~0 ~-0.9536529930730311 0 0 0 0 1 normal
particle flame ~-0.9518411972923277 ~0 ~4.136086751781068 0 0 0 0 1 normal
particle flame ~0.9050065201166948 ~0 ~-0.14250289336927713 0 0 0 0 1 normal
particle flame ~2.4930984759854002 ~0 ~-3.2040181234411804 0 0 0 0 1 normal
particle flame ~-2.0062642811395204 ~0 ~0.5779883441883715 0 0 0 0 1 normal
particle flame ~-2.980027687770848 ~0 ~1.593840224406417 0 0 0 0 1 normal
particle flame ~3.2076680149238372 ~0 ~0.16692524189707414 0 0 0 0 1 normal
particle flame ~2.286111001924712 ~0 ~-0.3105910864647832 0 0 0 0 1 normal
particle flame ~-3.552003432611318 ~0 ~-1.7786769051261813 0 0 0 0 1 normal
particle flame ~-1.0851352026663852 ~0 ~0.1041994984021768 0 0 0 0 1 normal
particle flame ~2.6484394346284867 ~0 ~3.323571760186601 0 0 0 0 1 normal
particle flame ~0.38058605267858114 ~0 ~-0.9527817372954772 0 0 0 0 1 normal
particle flame ~-0.9463101965623535 ~0 ~-3.890904891471331 0 0 0 0 1 normal
particle flame ~-0.8070872379847582 ~0 ~2.081100252691378 0 0 0 0 1 normal
particle flame ~-0.5726101744794936 ~0 ~3.2206266048916334 0 0 0 0 1 normal