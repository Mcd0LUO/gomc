scoreboard players tag @s remove EasyCoord.Main
scoreboard players tag @s remove EasyCoord.Nether
scoreboard players tag @s remove EasyCoord.End
scoreboard players tag @s add EasyCoord.Main {Dimension:0}
scoreboard players tag @s add EasyCoord.End {Dimension:1}
scoreboard players tag @s add EasyCoord.Nether {Dimension:-1}
scoreboard players tag @s add EasyCoord.HasLogged
playsound minecraft:entity.arrow.hit_player ambient @s
##记录x
summon minecraft:area_effect_cloud ~ ~ ~ {Tags:["EasyCoord.X_Marker"],PersistenceRequired:1b,Age:0,Duration:2000,Radius:0}

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-33554432 ~ ~ scoreboard players add @s EasyCoord:x 33554432
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-33554432 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-16777216 ~ ~ scoreboard players add @s EasyCoord:x 16777216
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-16777216 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-8388608 ~ ~ scoreboard players add @s EasyCoord:x 8388608
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-8388608 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-4194304 ~ ~ scoreboard players add @s EasyCoord:x 4194304
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-4194304 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-2097152 ~ ~ scoreboard players add @s EasyCoord:x 2097152
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-2097152 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-1048576 ~ ~ scoreboard players add @s EasyCoord:x 1048576
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-1048576 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-524288 ~ ~ scoreboard players add @s EasyCoord:x 524288
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-524288 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-262144 ~ ~ scoreboard players add @s EasyCoord:x 262144
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-262144 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-131072 ~ ~ scoreboard players add @s EasyCoord:x 131072
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-131072 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-65536 ~ ~ scoreboard players add @s EasyCoord:x 65536
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-65536 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-32768 ~ ~ scoreboard players add @s EasyCoord:x 32768
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-32768 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-16384 ~ ~ scoreboard players add @s EasyCoord:x 16384
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-16384 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-8192 ~ ~ scoreboard players add @s EasyCoord:x 8192
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-8192 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-4096 ~ ~ scoreboard players add @s EasyCoord:x 4096
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-4096 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-2048 ~ ~ scoreboard players add @s EasyCoord:x 2048
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-2048 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-1024 ~ ~ scoreboard players add @s EasyCoord:x 1024
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-1024 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-512 ~ ~ scoreboard players add @s EasyCoord:x 512
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-512 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-256 ~ ~ scoreboard players add @s EasyCoord:x 256
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-256 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-128 ~ ~ scoreboard players add @s EasyCoord:x 128
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-128 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-64 ~ ~ scoreboard players add @s EasyCoord:x 64
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-64 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-32 ~ ~ scoreboard players add @s EasyCoord:x 32
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-32 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-16 ~ ~ scoreboard players add @s EasyCoord:x 16
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-16 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-8 ~ ~ scoreboard players add @s EasyCoord:x 8
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-8 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-4 ~ ~ scoreboard players add @s EasyCoord:x 4
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-4 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-2 ~ ~ scoreboard players add @s EasyCoord:x 2
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-2 ~ ~

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-1 ~ ~ scoreboard players add @s EasyCoord:x 1
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] ~-1 ~ ~

scoreboard players remove @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] EasyCoord:x 30000000
scoreboard players operation @s EasyCoord:x = @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker] EasyCoord:x 
kill @e[type=area_effect_cloud,c=1,tag=EasyCoord.X_Marker]


##记录y
summon minecraft:area_effect_cloud ~ ~ ~ {Tags:["EasyCoord.Y_Marker"],PersistenceRequired:1b,Age:0,Duration:1200,Radius:0}

scoreboard players add @e[type=area_effect_cloud,c=1,tag=EasyCoord.Y_Marker] EasyCoord:y 0
execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Y_Marker] ~ ~ ~ stats entity @s set AffectedBlocks @s EasyCoord:y
execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Y_Marker] ~ ~ ~ testforblocks ~ 0 ~ ~ ~-1 ~ ~ 0 ~
scoreboard players operation @s EasyCoord:y = @e[type=area_effect_cloud,c=1,tag=EasyCoord.Y_Marker] EasyCoord:y
kill @e[type=area_effect_cloud,c=1,tag=EasyCoord.Y_Marker]


##记录z
summon minecraft:area_effect_cloud ~ ~ ~ {Tags:["EasyCoord.Z_Marker"],PersistenceRequired:1b,Potion:CBC,Age:0,Duration:1200,Radius:0}

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-33554432 scoreboard players add @s EasyCoord:z 33554432
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-33554432

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-16777216 scoreboard players add @s EasyCoord:z 16777216
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-16777216

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-8388608 scoreboard players add @s EasyCoord:z 8388608
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-8388608

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-4194304 scoreboard players add @s EasyCoord:z 4194304
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-4194304

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-2097152 scoreboard players add @s EasyCoord:z 2097152
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-2097152

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-1048576 scoreboard players add @s EasyCoord:z 1048576
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-1048576

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-524288 scoreboard players add @s EasyCoord:z 524288
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-524288

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-262144 scoreboard players add @s EasyCoord:z 262144
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-262144

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-131072 scoreboard players add @s EasyCoord:z 131072
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-131072

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-65536 scoreboard players add @s EasyCoord:z 65536
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-65536

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-32768 scoreboard players add @s EasyCoord:z 32768
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-32768

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-16384 scoreboard players add @s EasyCoord:z 16384
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-16384

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-8192 scoreboard players add @s EasyCoord:z 8192
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-8192

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-4096 scoreboard players add @s EasyCoord:z 4096
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-4096

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-2048 scoreboard players add @s EasyCoord:z 2048
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-2048

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-1024 scoreboard players add @s EasyCoord:z 1024
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-1024

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-512 scoreboard players add @s EasyCoord:z 512
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-512

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-256 scoreboard players add @s EasyCoord:z 256
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-256

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-128 scoreboard players add @s EasyCoord:z 128
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-128

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-64 scoreboard players add @s EasyCoord:z 64
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-64

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-32 scoreboard players add @s EasyCoord:z 32
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-32

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-16 scoreboard players add @s EasyCoord:z 16
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-16

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-8 scoreboard players add @s EasyCoord:z 8
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-8

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-4 scoreboard players add @s EasyCoord:z 4
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-4

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-2 scoreboard players add @s EasyCoord:z 2
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-2

execute @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-1 scoreboard players add @s EasyCoord:z 1
tp @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] ~ ~ ~-1

scoreboard players remove @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] EasyCoord:z 30000000
scoreboard players operation @s EasyCoord:z = @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker] EasyCoord:z
kill @e[type=area_effect_cloud,c=1,tag=EasyCoord.Z_Marker]