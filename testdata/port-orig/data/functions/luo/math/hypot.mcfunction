#@desc : 用于计算直角三角形斜边
#@param arg1: 直角边1
#@param arg2: 直角边2
#@return return: 斜边

##
    scoreboard players operation @s arg1 *= @s arg1 
    scoreboard players operation @s arg2 *= @s arg2
    scoreboard players operation @s arg1 += @s arg2
    function luo:math/sqrt/NIM

#