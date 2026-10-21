function sys:menu/chest/材料/末法点/fast_ini unless @s[score_math.operation1_min=1]

give @s[score_math.operation1_min=64] paper 64 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1末法点§f⌭┓"},Tags:["货币","末法点","1","undrop"]}
scoreboard players remove @s[score_math.operation1_min=64] math.operation1 64
give @s[score_math.operation1_min=32] paper 32 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1末法点§f⌭┓"},Tags:["货币","末法点","1","undrop"]}
scoreboard players remove @s[score_math.operation1_min=32] math.operation1 32
give @s[score_math.operation1_min=16] paper 16 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1末法点§f⌭┓"},Tags:["货币","末法点","1","undrop"]}
scoreboard players remove @s[score_math.operation1_min=16] math.operation1 16
give @s[score_math.operation1_min=8] paper 8 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1末法点§f⌭┓"},Tags:["货币","末法点","1","undrop"]}
scoreboard players remove @s[score_math.operation1_min=8] math.operation1 8
give @s[score_math.operation1_min=4] paper 4 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1末法点§f⌭┓"},Tags:["货币","末法点","1","undrop"]}
scoreboard players remove @s[score_math.operation1_min=4] math.operation1 4
give @s[score_math.operation1_min=2] paper 2 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1末法点§f⌭┓"},Tags:["货币","末法点","1","undrop"]}
scoreboard players remove @s[score_math.operation1_min=2] math.operation1 2
give @s[score_math.operation1_min=1] paper 1 0 {display:{Lore:["§6⛂货币"],Name:"§f┏⌭§e1末法点§f⌭┓"},Tags:["货币","末法点","1","undrop"]}
scoreboard players remove @s[score_math.operation1_min=1] math.operation1 1

function sys:menu/chest/材料/末法点/1_fast_pass unless @s[score_math.operation1=0] 