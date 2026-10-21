# 牛顿迭代法开平方
  #参数: (arg1) 待开方数  返回值: (return) 结果

  #预测值: 传入值/4
  scoreboard players operation @s temp1 = @s arg1
  scoreboard players operation @s temp1 /= math.4 math
  scoreboard players operation @s temp2 = @s arg1
  scoreboard players operation @s temp4 = @s temp1
  #迭代逼近
  function luo:math/sqrt/__iteration
  #迭代结束
    # 返回结果
    scoreboard players operation @s return = @s temp1
    #reset
    scoreboard players reset @s temp1
    scoreboard players reset @s temp2
    scoreboard players reset @s temp3
    scoreboard players reset @s temp4
    scoreboard players reset @s arg1
    
#end