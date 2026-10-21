give @s[score_money_min=100] paper 1 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e100末法点§f⌭┓"},Tags:["货币","末法点","100","undrop"]}
tellraw @s[score_money=99] ["§r[§eBANK§r]:§7虚拟末法点不足"]
tp @s[score_money=99] ~ ~ ~0.5
scoreboard players remove @s[score_money_min=100] money 100