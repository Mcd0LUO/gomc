scoreboard players operation @s math.operation1 = @s money
scoreboard players remove @s math.operation1 850
scoreboard players remove @s[score_math.operation1_min=0] money 850
tellraw @s[score_math.operation1_min=0] ["§r[§eSHOP§r]:§7购买成功,剩余末法点 ",{"score":{"objective":"money","name":"@s"},"color":"red"}]
give @s[score_math.operation1_min=0] minecraft:dye 1 4 {display:{Lore:["§b༖§7珍贵材料"],Name:"§f┏⌭§b灵§f⌭┓"},Tags:["材料","灵"]}

tellraw @s[score_math.operation1=-1] ["§r[§eSHOP§r]:§7购买失败,缺少末法点 ",{"score":{"objective":"math.operation1","name":"@s"},"color":"red"},"§c-"]
scoreboard players reset @s math.operation1
function sys:menu/chest/材料/update