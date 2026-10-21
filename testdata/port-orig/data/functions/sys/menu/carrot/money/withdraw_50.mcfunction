give @s[score_money_min=50] paper 1 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e50末法点§f⌭┓"},Tags:["货币","末法点","50","undrop"]}
tellraw @s[score_money=49] ["§r[§eBANK§r]:§7虚拟末法点不足"]
tp @s[score_money=49] ~ ~ ~0.5
scoreboard players remove @s[score_money_min=50] money 50