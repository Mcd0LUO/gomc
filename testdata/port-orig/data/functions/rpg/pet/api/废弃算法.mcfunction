# # 8方向盔甲架
# summon armor_stand ~ ~1 ~ {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# summon armor_stand ~ ~-1 ~ {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}

# summon armor_stand ~1 ~ ~ {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# summon armor_stand ~-1 ~ ~ {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# summon armor_stand ~ ~ ~1 {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# summon armor_stand ~ ~ ~-1 {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}

# summon armor_stand ~0.7071067811865476 ~ ~0.7071067811865476 {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# summon armor_stand ~0.7071067811865476 ~ ~-0.7071067811865476 {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# summon armor_stand ~-0.7071067811865476 ~ ~0.7071067811865476 {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# summon armor_stand ~-0.7071067811865476 ~ ~-0.7071067811865476 {Small:1b,NoAI:1b,NoGravity:1b,Marker:1b,Tags:["pet.path"],Invisible:1b}
# # 选取距离mob最近盔甲架
# execute @e[tag=mob.asg,c=1,r=25] ~ ~ ~ scoreboard players tag @e[c=1,tag=pet.path] add pet.nearest
# # tp最近盔甲架
# tp @s @e[c=1,tag=pet.nearest]
# # 释放资源
# kill @e[r=3,tag=pet.path]