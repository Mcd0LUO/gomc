scoreboard players reset @s elemet.saveTotal
entitydata @s {CustomName:""}
execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s arg1 *= 2 num

execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s hurt.armor * @s 盾伤倍率

execute @e[tag=mob.asg,dy=-2] ~ ~ ~ scoreboard players operation @s armor.all.now -= @s hurt.armor

particle largeexplode ~ ~-1 ~ 0 0 0 0.25 2 normal
particle fireworksSpark ~ ~ ~ 0 0 0 0.25 10 normal

summon lightning_bolt ~ ~4.3 ~

execute @a[score_temp.uid.match_min=0,score_temp.uid.match=0] ~ ~ ~ function rpg:monster/hurt/normal/thunder/backToPlayer

tellraw @a[score_temp.uid.match_min=0,score_temp.uid.match=0] [["§7[§e§l⚡ §r| §c§l༗§7]->超载 剧烈爆炸 元素溅射"]]
scoreboard players tag @s remove elemet.saveFire

particle dragonbreath ~0.0 ~ ~-2.5 0.0 0 -1.0 0.3 0 normal
particle dragonbreath ~-0.75 ~0.1 ~0.0 0 0 0 0 1 normal
particle dragonbreath ~-0.9050999681267564 ~0.1 ~4.085007963455079 0 0 0 0 1 normal
particle dragonbreath ~1.3311049049897925 ~0.1 ~-0.9716043466392063 0 0 0 0 1 normal
particle dragonbreath ~2.20728750021509 ~0.1 ~-2.932802416010863 0 0 0 0 1 normal
particle dragonbreath ~-2.6550507757909885 ~0.1 ~1.0328381860936036 0 0 0 0 1 normal
particle dragonbreath ~-2.334963470471356 ~0.1 ~1.39296299788045 0 0 0 0 1 normal
particle dragonbreath ~3.7545700824935535 ~0.1 ~0.0912936066036304 0 0 0 0 1 normal
particle dragonbreath ~1.4281324518536902 ~0.1 ~-0.465676527463152 0 0 0 0 1 normal
particle dragonbreath ~-3.7855389873971537 ~0.1 ~-1.8452630991649515 0 0 0 0 1 normal
particle dragonbreath ~-0.33174741884758974 ~0.1 ~0.6952333845286192 0 0 0 0 1 normal
particle dragonbreath ~2.6133633903204894 ~0.1 ~3.2233737891661405 0 0 0 0 1 normal
particle dragonbreath ~0.006061197566168852 ~0.1 ~-1.7949531471810145 0 0 0 0 1 normal
particle dragonbreath ~-0.9084785077704557 ~0.1 ~-3.462230250440782 0 0 0 0 1 normal
particle dragonbreath ~-0.8574286104597395 ~0.1 ~2.8483269626383074 0 0 0 0 1 normal
particle dragonbreath ~-0.31523711548807065 ~0.1 ~2.565724225878911 0 0 0 0 1 normal
particle dragonbreath ~2.453201637537061 ~0.1 ~-2.9469916233137816 0 0 0 0 1 normal
particle dragonbreath ~0.39944336254390045 ~0.1 ~-1.298949212843601 0 0 0 0 1 normal
particle dragonbreath ~-3.818285761032864 ~0.1 ~1.8235405857654625 0 0 0 0 1 normal
particle dragonbreath ~0.5064105649866323 ~0.1 ~0.6553268274858048 0 0 0 0 1 normal
particle dragonbreath ~4.10799206638441 ~0.1 ~-0.04074408772556681 0 0 0 0 1 normal
particle dragonbreath ~-1.5527702326104182 ~0.1 ~-1.166131029393294 0 0 0 0 1 normal
particle dragonbreath ~-3.191144070122427 ~0.1 ~-1.393976569622992 0 0 0 0 1 normal
particle dragonbreath ~1.7829176581911084 ~0.1 ~2.5307580192026418 0 0 0 0 1 normal
particle dragonbreath ~1.7395541650734936 ~0.1 ~1.7218483036477275 0 0 0 0 1 normal
particle dragonbreath ~-0.8003433700025877 ~0.1 ~-3.823892790924206 0 0 0 0 1 normal
particle dragonbreath ~-0.7619160521640718 ~0.1 ~-0.9536529930730311 0 0 0 0 1 normal
particle dragonbreath ~-0.9518411972923277 ~0.1 ~4.136086751781068 0 0 0 0 1 normal
particle dragonbreath ~0.9050065201166948 ~0.1 ~-0.14250289336927713 0 0 0 0 1 normal
particle dragonbreath ~2.4930984759854002 ~0.1 ~-3.2040181234411804 0 0 0 0 1 normal
particle dragonbreath ~-2.0062642811395204 ~0.1 ~0.5779883441883715 0 0 0 0 1 normal
particle dragonbreath ~-2.980027687770848 ~0.1 ~1.593840224406417 0 0 0 0 1 normal
particle dragonbreath ~3.2076680149238372 ~0.1 ~0.16692524189707414 0 0 0 0 1 normal
particle dragonbreath ~2.286111001924712 ~0.1 ~-0.3105910864647832 0 0 0 0 1 normal
particle dragonbreath ~-3.552003432611318 ~0.1 ~-1.7786769051261813 0 0 0 0 1 normal
particle dragonbreath ~-1.0851352026663852 ~0.1 ~0.1041994984021768 0 0 0 0 1 normal
particle dragonbreath ~2.6484394346284867 ~0.1 ~3.323571760186601 0 0 0 0 1 normal
particle dragonbreath ~0.38058605267858114 ~0.1 ~-0.9527817372954772 0 0 0 0 1 normal
particle dragonbreath ~-0.9463101965623535 ~0.1 ~-3.890904891471331 0 0 0 0 1 normal
particle dragonbreath ~-0.8070872379847582 ~0.1 ~2.081100252691378 0 0 0 0 1 normal
particle dragonbreath ~-0.5726101744794936 ~0.1 ~3.2206266048916334 0 0 0 0 1 normal
particle dragonbreath ~2.1878547743195575 ~0.1 ~-2.501331697121448 0 0 0 0 1 normal
particle dragonbreath ~1.07024587792804 ~0.1 ~-1.8808211895323739 0 0 0 0 1 normal
particle dragonbreath ~-3.6547825712342172 ~0.1 ~1.7014723108852787 0 0 0 0 1 normal
particle dragonbreath ~-0.4284800118701563 ~0.1 ~0.8721106461459645 0 0 0 0 1 normal
particle dragonbreath ~4.248021886382343 ~0.1 ~-0.004430855276010795 0 0 0 0 1 normal
particle dragonbreath ~-0.6735597765714504 ~0.1 ~-0.9330595387856111 0 0 0 0 1 normal
particle dragonbreath ~-3.579753733148772 ~0.1 ~-1.6510276562856263 0 0 0 0 1 normal
particle dragonbreath ~1.2346365618920294 ~0.1 ~2.027764942913337 0 0 0 0 1 normal
particle dragonbreath ~2.1023731844253093 ~0.1 ~2.355836961415194 0 0 0 0 1 normal
particle dragonbreath ~-0.6297340998499353 ~0.1 ~-3.370600421466609 0 0 0 0 1 normal
particle dragonbreath ~-0.7953201478084264 ~0.1 ~-1.855593960465654 0 0 0 0 1 normal
particle dragonbreath ~-0.952255875210935 ~0.1 ~3.9709760916256376 0 0 0 0 1 normal
particle dragonbreath ~0.48748871429784746 ~0.1 ~0.7221069344344089 0 0 0 0 1 normal
particle dragonbreath ~2.634775272757495 ~0.1 ~-3.3207415934196938 0 0 0 0 1 normal
particle dragonbreath ~-1.2845731589618068 ~0.1 ~0.04721239879498906 0 0 0 0 1 normal
particle dragonbreath ~-3.4564131966846166 ~0.1 ~1.7480349612340433 0 0 0 0 1 normal
particle dragonbreath ~2.49800467174583 ~0.1 ~0.27548349254127213 0 0 0 0 1 normal
particle dragonbreath ~3.032802641093289 ~0.1 ~-0.19242481003146317 0 0 0 0 1 normal
particle dragonbreath ~-3.124662169263684 ~0.1 ~-1.6403871531895555 0 0 0 0 1 normal
particle dragonbreath ~-1.8191778046648275 ~0.1 ~-0.4429412758227387 0 0 0 0 1 normal
particle dragonbreath ~2.5450007387444087 ~0.1 ~3.251061231372945 0 0 0 0 1 normal
particle dragonbreath ~0.7917892499800696 ~0.1 ~-0.0865328892688868 0 0 0 0 1 normal
particle dragonbreath ~-0.9559301118451606 ~0.1 ~-4.113157580230932 0 0 0 0 1 normal
particle dragonbreath ~-0.768929506177066 ~0.1 ~1.201370728192228 0 0 0 0 1 normal
particle dragonbreath ~-0.7614915651325791 ~0.1 ~3.720785356963893 0 0 0 0 1 normal
particle dragonbreath ~1.841957646790385 ~0.1 ~-1.9036909725565903 0 0 0 0 1 normal
particle dragonbreath ~1.6479804257871373 ~0.1 ~-2.405120807960282 0 0 0 0 1 normal
particle dragonbreath ~-3.310502083811654 ~0.1 ~1.4729685118467313 0 0 0 0 1 normal
particle dragonbreath ~-1.3259236783597277 ~0.1 ~1.1038088666309605 0 0 0 0 1 normal
particle dragonbreath ~4.166757684112818 ~0.1 ~0.030164779655438234 0 0 0 0 1 normal
particle dragonbreath ~0.2562238685830342 ~0.1 ~-0.7109188048231667 0 0 0 0 1 normal
particle dragonbreath ~-3.7934156141475617 ~0.1 ~-1.8011622467109683 0 0 0 0 1 normal
particle dragonbreath ~0.5857339573890707 ~0.1 ~1.4577584742221412 0 0 0 0 1 normal
particle dragonbreath ~2.3917648005499847 ~0.1 ~2.844144774766595 0 0 0 0 1 normal
particle dragonbreath ~-0.3904352429375386 ~0.1 ~-2.753462061594005 0 0 0 0 1 normal
particle dragonbreath ~-0.8434039806628659 ~0.1 ~-2.6571372245168954 0 0 0 0 1 normal
particle dragonbreath ~-0.9204740293385717 ~0.1 ~3.5957618882637847 0 0 0 0 1 normal
particle dragonbreath ~0.10133520352814873 ~0.1 ~1.5753794627644657 0 0 0 0 1 normal
particle dragonbreath ~2.635634505483527 ~0.1 ~-3.2673679352643394 0 0 0 0 1 normal
particle dragonbreath ~-0.5322002910831931 ~0.1 ~-0.5349843957282494 0 0 0 0 1 normal


particle sweepAttack ~0.0 ~ ~-2.0 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.147153 ~ ~-1.638304 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.879385 ~ ~-0.68404 -0.0 0 0.0 0.3 1 normal
particle sweepAttack ~1.931852 ~ ~0.517638 -0.0 0 -0.0 0.3 1 normal