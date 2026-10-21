scoreboard players reset @s armor.fire
scoreboard players reset @s armor.water
scoreboard players reset @s armor.ice
scoreboard players reset @s armor.thunder

scoreboard players operation @s armor.fire = @s armor.fire.tr
scoreboard players operation @s armor.water = @s armor.water.tr
scoreboard players operation @s armor.ice = @s armor.ice.tr
scoreboard players operation @s armor.thunder = @s armor.thunder.tr

scoreboard players set @s armor.putonStep 1
tellraw @s ["§r[§e护甲系统§r] §d头盔§7属性重载成功,允许重载§d胸甲属性!"]