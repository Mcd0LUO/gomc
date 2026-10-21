## 普通武器攻击主函数
    # 获取主手武器属性
        scoreboard players set @s element.type 1 {SelectedItem:{tag:{Tags:["element.fire"]}}}
        scoreboard players set @s element.type 2 {SelectedItem:{tag:{Tags:["element.water"]}}} 
        scoreboard players set @s element.type 3 {SelectedItem:{tag:{Tags:["element.ice"]}}} 
        scoreboard players set @s element.type 4 {SelectedItem:{tag:{Tags:["element.thunder"]}}} 
        scoreboard players set @s element.type 5 {SelectedItem:{tag:{Tags:["element.leaf"]}}}
    # 标记
        scoreboard players tag @s add self.atk 
        scoreboard players tag @e[r=15,tag=mob] add mob.hurt {HurtTime:10s}
        scoreboard players operation @e[r=15,tag=mob.hurt] uid = @s uid 
    # 分发到具体的函数
        function rpg:player/attack/normal_weapon/fire if @s[score_element.type=1,score_element.type_min=1]
        function rpg:player/attack/normal_weapon/water if @s[score_element.type=2,score_element.type_min=2]

    # 清理
        advancement revoke @s only event:OnPlayerHurtMonster
        scoreboard players set @s §cevent 0
#