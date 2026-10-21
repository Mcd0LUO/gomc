tp @a[score_menu.trigger_min=4001,score_menu.trigger=4001] 65 11 -163 90 0
title @a[score_menu.trigger_min=4001,score_menu.trigger=4001] title ["§e副本传送"]

tp @a[score_menu.trigger_min=4002,score_menu.trigger=4002] 123.3 31 -158.7 -40 0
title @a[score_menu.trigger_min=4002,score_menu.trigger=4002] title ["§e兑换塔楼"]

tp @a[score_menu.trigger_min=4003,score_menu.trigger=4003] 101.3 9 -149.7 0 0
title @a[score_menu.trigger_min=4003,score_menu.trigger=4003] title ["§e元素魔屋"]

tp @a[score_menu.trigger_min=4004,score_menu.trigger=4004] 59 14 -218
title @a[score_menu.trigger_min=4004,score_menu.trigger=4004] title ["§e星升之铺"]

tp @a[score_menu.trigger_min=4005,score_menu.trigger=4005] 80 1 -150 44 0
title @a[score_menu.trigger_min=4005,score_menu.trigger=4005] title ["§e功能屋"]
title @a[score_menu.trigger_min=4005,score_menu.trigger=4005] subtitle ["§7银行/扭蛋/额外背包使用 施工中ing"]


playsound item.chorus_fruit.teleport master @a[score_menu.trigger_min=4001,score_menu.trigger=4999] ~0 ~0 ~0 3 1 0

execute @a[score_menu.trigger_min=4001,score_menu.trigger=4009,tag=释放宠物] ~ ~ ~ function sys:menu/carrot/pet/sum_list/召回