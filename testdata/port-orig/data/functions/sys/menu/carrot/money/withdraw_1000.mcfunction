give @s[score_money_min=1000] paper 1 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1000末法点§f⌭┓"},Tags:["货币","末法点","1000","undrop"]}
tellraw @s[score_money=999] ["§r[§eBANK§r]:§7虚拟末法点不足"]
tp @s[score_money=999] ~ ~ ~0.5
scoreboard players remove @s[score_money_min=1000] money 1000