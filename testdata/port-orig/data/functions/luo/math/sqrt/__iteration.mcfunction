# 迭代部分 求解 ( f(x) = x^2 - arg = 0 )
  #计算新x值 -> x_next
  scoreboard players operation @s temp2 = @s arg1
  scoreboard players operation @s temp2 /= @s temp1
  scoreboard players operation @s temp1 += @s temp2
  scoreboard players operation @s temp1 /= math.2 math
  # 此时temp1为x_next 
  
  #判断是否收敛
  scoreboard players operation @s temp3 = @s temp1
  scoreboard players operation @s temp3 -= @s temp4
  # if x_next - x ∈ [-1,1]:

  # 继续迭代
  scoreboard players operation @s temp4 = @s temp1
  function luo:math/sqrt/__iteration unless @s[score_temp3_min=-1,score_temp3=1]
#end