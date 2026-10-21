particle dragonbreath ~0.0 ~1 ~-3.5 -0.0 0 1.0 0.3 0 normal
particle dragonbreath ~0.905867 ~1 ~-3.38074 -0.258819 0 0.965926 0.3 0 normal
particle dragonbreath ~1.75 ~1 ~-3.031089 -0.5 0 0.866025 0.3 0 normal
particle dragonbreath ~2.474874 ~1 ~-2.474874 -0.707107 0 0.707107 0.3 0 normal
particle dragonbreath ~3.031089 ~1 ~-1.75 -0.866025 0 0.5 0.3 0 normal
particle dragonbreath ~3.38074 ~1 ~-0.905867 -0.965926 0 0.258819 0.3 0 normal
particle dragonbreath ~3.5 ~1 ~0.0 -1.0 0 -0.0 0.3 0 normal
particle dragonbreath ~3.38074 ~1 ~0.905867 -0.965926 0 -0.258819 0.3 0 normal
particle dragonbreath ~3.031089 ~1 ~1.75 -0.866025 0 -0.5 0.3 0 normal
particle dragonbreath ~2.474874 ~1 ~2.474874 -0.707107 0 -0.707107 0.3 0 normal
particle dragonbreath ~1.75 ~1 ~3.031089 -0.5 0 -0.866025 0.3 0 normal
particle dragonbreath ~0.905867 ~1 ~3.38074 -0.258819 0 -0.965926 0.3 0 normal
particle dragonbreath ~-0.0 ~1 ~3.5 0.0 0 -1.0 0.3 0 normal
particle dragonbreath ~-0.905867 ~1 ~3.38074 0.258819 0 -0.965926 0.3 0 normal
particle dragonbreath ~-1.75 ~1 ~3.031089 0.5 0 -0.866025 0.3 0 normal
particle dragonbreath ~-2.474874 ~1 ~2.474874 0.707107 0 -0.707107 0.3 0 normal
particle dragonbreath ~-3.031089 ~1 ~1.75 0.866025 0 -0.5 0.3 0 normal
particle dragonbreath ~-3.38074 ~1 ~0.905867 0.965926 0 -0.258819 0.3 0 normal
particle dragonbreath ~-3.5 ~1 ~-0.0 1.0 0 0.0 0.3 0 normal
particle dragonbreath ~-3.38074 ~1 ~-0.905867 0.965926 0 0.258819 0.3 0 normal
particle dragonbreath ~-3.031089 ~1 ~-1.75 0.866025 0 0.5 0.3 0 normal
particle dragonbreath ~-2.474874 ~1 ~-2.474874 0.707107 0 0.707107 0.3 0 normal
particle dragonbreath ~-1.75 ~1 ~-3.031089 0.5 0 0.866025 0.3 0 normal
particle dragonbreath ~-0.905867 ~1 ~-3.38074 0.258819 0 0.965926 0.3 0 normal

function rpg:monster/hurt/elemet/thunder/main_thunder

scoreboard players tag @s remove mob.target

effect @s minecraft:instant_damage 1 1
effect @s minecraft:instant_health 1 1