## 通用首调函数
    # 懒加载赋值属性
    function rpg:sum/asg/asg_main unless @s[tag=mob.asg]
    # 获取攻击玩家
    scoreboard players operation @a[tag=self.atk,r=12] temp.uid.match = @s uid
    execute @a[tag=self.atk,r=12] ~ ~ ~ scoreboard players operation @s temp.uid.match -= @s uid

#