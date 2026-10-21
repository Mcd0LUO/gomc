summon minecraft:armor_stand ~ ~1.5 ~ {CustomName:"§e§l[MISS]",Tags:["Display","文显","new"],CustomNameVisible:1b,Invisible:1,Small:1,Marker:1b}

execute @e[c=1,tag=文显] ~ ~ ~ function sys:api/文显/main if @s[tag=new]