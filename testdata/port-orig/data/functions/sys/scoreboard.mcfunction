# all记分板save

# 通用事件触发器
scoreboard objectives add §cevent trigger 通用事件触发器

scoreboard objectives add ele.type dummy 元素类型

#### sys
    scoreboard objectives add temp.uid.match dummy uid匹配临时值

#
# 元素武器分选id
scoreboard objectives add weapon.eleid dummy 元素武器分选id
# 武器元素杆触发
scoreboard objectives add weapon.use stat.useItem.minecraft.carrot_on_a_stick 武器元素杆触发
# 武器使用次数
scoreboard objectives add weapon.useTotal dummy 武器使用次数
# 血量
scoreboard objectives add health dummy
scoreboard objectives add health.max dummy
scoreboard objectives add health.restore dummy 回复血量3s
# 振刀时间戳
scoreboard objectives add damage.time dummy

scoreboard objectives add ele.type dummy 攻击附带的元素属性

# 攻击值总计-----------
scoreboard objectives add damage.total dummy
# 元素挂留印记总计
scoreboard objectives add ele.saveTotal dummy
# 火元素攻击
scoreboard objectives add damage.fire trigger 
# 水元素攻击
scoreboard objectives add damage.water dummy
# 雷元素攻击
scoreboard objectives add damage.thunder dummy 
# 冰元素攻击
scoreboard objectives add damage.ice dummy
# 进度触发伤害赋值
scoreboard objectives add damage.advset trigger
# 进度触发伤害类型赋值
scoreboard objectives add damage.eleid trigger
# 元素护盾
scoreboard objectives add armor.all.now dummy
# 元素护盾
scoreboard objectives add armor.all.max dummy
# 玩家盾伤
scoreboard objectives add damage.armor dummy
# 玩家魔抗
scoreboard objectives add armor.player.all dummy
# 真伤
scoreboard objectives add damage.real dummy
# 怒气值
scoreboard objectives add angry dummy
# 元素挂留id
scoreboard objectives add ele.save.id dummy
# 玩家元素等级
scoreboard objectives add ele.savetime dummy
scoreboard objectives add ele.fire_lv dummy
scoreboard objectives add ele.water_lv dummy
scoreboard objectives add ele.ice_lv dummy
scoreboard objectives add ele.thund_lv dummy
scoreboard objectives add ele.armor_lv dummy
# 受击值总计
scoreboard objectives add hurt.total dummy
scoreboard objectives add hurt.fire dummy
scoreboard objectives add hurt.water dummy
scoreboard objectives add hurt.thunder dummy
scoreboard objectives add hurt.ice dummy
scoreboard objectives add hurt.armor dummy
scoreboard objectives add hurt.real dummy
# 额外元素攻击值
scoreboard objectives add more.fire dummy
scoreboard objectives add more.water dummy
scoreboard objectives add more.ice dummy
scoreboard objectives add more.thunder dummy
# 受击零时值
scoreboard objectives add hurt.temp dummy
# 受击时间戳
scoreboard objectives add hurt.time dummy
# -------------

# 蓝量
scoreboard objectives add mp.max dummy
scoreboard objectives add mp.now dummy
scoreboard objectives add mp.MoreRestore dummy 额外蓝量恢复
# 菜单使用
scoreboard objectives add menu.use stat.useItem.minecraft.carrot_on_a_stick
scoreboard objectives add menu.trigger trigger
# cd计时
scoreboard objectives add cd.1 dummy
# 功能区域-------------
# uid
scoreboard objectives add uid dummy
# uid_ini
function sys:basic/初始化/uid/ini
scoreboard objectives add frequency dummy
# 进度
scoreboard objectives add adv trigger
# 离开游戏
scoreboard objectives add leaveGame dummy
# 打开箱子
scoreboard objectives add chest.open stat.chestOpened
scoreboard objectives add shift.time stat.sneakTime 潜行时间
scoreboard objectives add stats.item dummy
# 行走距离
scoreboard objectives add data.walk stat.walkOneCm
# uid匹配临时值
scoreboard objectives add temp.uid.match dummy
# 末法点
scoreboard objectives add money dummy 末法点
scoreboard objectives add dq dummy 浩瀚星核
# 队伍系统
# 队伍id
scoreboard objectives add team.id dummy
# 队伍加入输入
scoreboard objectives add team.join trigger
# 队伍零时密钥
scoreboard objectives add tema.key dummy
# 队伍临时值
scoreboard objectives add team.temp dummy
# 剧情进度
scoreboard objectives add plot.id dummy 剧情大进度id
scoreboard objectives add plot.time dummy 剧情计时
scoreboard objectives add plot.step dummy 剧情小进度

scoreboard objectives add adv.glfd dummy 孤落法典进度
# 刷新计时
scoreboard objectives add sum.time dummy 刷新计时
scoreboard objectives add sum.nearByPlayer dummy 刷新统计玩家
scoreboard objectives add sum.nearByMob dummy 刷新统计怪物
# 抽奖
scoreboard objectives add talk stat.talkedToVillager 村民对话
# 护甲穿戴监测
scoreboard objectives add §7armor.water trigger 水元素抵抗
scoreboard objectives add §7armor.fire trigger 火元素抵抗
scoreboard objectives add §7armor.ice trigger 冰元素抵抗
scoreboard objectives add §7armor.thunder trigger 雷元素抵抗
scoreboard objectives add §7armor.test trigger 护甲监听优化
scoreboard objectives add armor.putonStep dummy 穿戴步骤

#  称号系统
scoreboard objectives add name.id dummy 称号id
scoreboard objectives add name.+HpMax dummy 称号血量上限
scoreboard objectives add name.+MpMax dummy 称号蓝量上限
scoreboard objectives add name.+MpRestore dummy 称号蓝量回复
scoreboard objectives add name.+HpRestore dummy 称号血量回复

scoreboard objectives add name.承包鱼塘 dummy 承包鱼塘等级
scoreboard objectives add name.save dummy
scoreboard teams add 承包鱼塘
gamerule randomTickSpeed 0
gamerule showDeathMessages false 
# 扩展属性
scoreboard objectives add data.kill dummy 击杀怪物数
scoreboard objectives add ifm.board dummy §7◈ §e§l⚶ §7◈

# 二本专属
scoreboard objectives add bar.cold dummy 寒冷值
scoreboard objectives add bar.cold.max dummy 最大寒冷值
scoreboard objectives add debuff.armor dummy 减元素抗性
# 技能专属
scoreboard objectives add skill.time.fire dummy 火技能计时
scoreboard objectives add skill.time.water dummy 水技能计时
scoreboard objectives add skill.time.ice dummy 冰技能计时
scoreboard objectives add skill.time.thund dummy 雷技能计时
scoreboard objectives add skill.temp2 dummy 技能临时变量2
scoreboard objectives add skill.id dummy 技能id
scoreboard objectives add skill.temp dummy 技能临时变量
scoreboard objectives add cd.f dummy 火cd
scoreboard objectives add cd.w 水cd
scoreboard objectives add cd.i 冰cd
scoreboard objectives add cd.t 雷cd
scoreboard objectives add skill.step dummy 技能步骤
scoreboard objectives add skill.time.clear dummy 技能飞行物计时
# 宠物
scoreboard objectives add pet.uid dummy 宠物uid
scoreboard objectives add pet.temp dummy
scoreboard objectives add pet.now dummy 拥有宠物id
scoreboard objectives add pet.lv dummy 宠物等级
scoreboard objectives add pet.time dummy 宠物计时
# 通用临时变量
scoreboard objectives add temp dummy 通用临时变量
scoreboard objectives add timing dummy 通用计时
# 实体uid
scoreboard objectives add uid.mob dummy 实体uid
# 坐标监测
scoreboard objectives add pos.x trigger
scoreboard objectives add pos.y trigger
scoreboard objectives add pos.z trigger