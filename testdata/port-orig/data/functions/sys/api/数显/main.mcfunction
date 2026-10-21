kill @e[tag=数显,dy=2]

summon armor_stand ~ ~2.45 ~ {Small:1b,Invisible:1b,NoGravity:1b,Marker:1b,Tags:["数显","数显.main"],Passengers:[{id:armor_stand,Small:1b,Invisible:1b,NoGravity:1b,Marker:1b,CustomName:"-",CustomNameVisible:1b,Tags:["数显","数显.4"]},{id:armor_stand,Small:1b,Invisible:1b,NoGravity:1b,Marker:1b,CustomName:"-",CustomNameVisible:1b,Tags:["数显","数显.3"]},{id:armor_stand,Small:1b,Invisible:1b,NoGravity:1b,Marker:1b,CustomName:"-",CustomNameVisible:1b,Tags:["数显","数显.2"]},{id:armor_stand,Small:1b,Invisible:1b,NoGravity:1b,Marker:1b,CustomName:"-",CustomNameVisible:1b,Tags:["数显","数显.1"]}]}

scoreboard players operation @e[c=1,tag=数显.main] arg1 = @s arg1
scoreboard players operation @e[c=1,tag=数显.main] temp = @s arg1
scoreboard players operation @e[c=1,tag=数显.main] temp %= 10 num
execute @e[tag=数显.main,c=1] ~ ~ ~ function sys:api/数显/asg_1
scoreboard players reset @e[c=1,tag=数显.main] temp
scoreboard players tag @e[c=1,tag=数显.main] remove 数显.main