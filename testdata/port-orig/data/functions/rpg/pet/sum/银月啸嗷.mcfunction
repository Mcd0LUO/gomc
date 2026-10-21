summon minecraft:wolf ~ ~ ~ {CustomName:"§r<§e⚔§r|§9§l☁§r> 银月§b啸嗷",CustomNameVisible:1b,PersistenceRequired:1b,Invulnerable:1b,Tags:["pet","pet.银月啸傲","pet.银月啸傲.new"]}

scoreboard players operation @e[c=1,tag=pet.银月啸傲.new] pet.uid = @s uid
scoreboard players tag @e[c=1,tag=pet.银月啸傲.new] remove pet.银月啸傲.new

tellraw @s ["§e✑ §7成功召唤§e⚔ §r银月§b啸嗷!"]
scoreboard players tag @s add 释放宠物