import math
from mcdev.pic import Rotate_3d

# 定义时间范围和步长
t_start = 0
t_end = 30
dt = 0.4

# 定义粒子轨迹的参数
scale = 0.17
rx = 0.0
ry = 0.0

# 计算粒子的运动轨迹
positions = []
for t in range(int((t_end - t_start) / dt)):
    time = t * dt
    x = ((0 + 5 * math.cos(time * -2) + 10 * math.cos(time * 1 + 3)) * scale) * math.cos(ry) + (((0) * scale) * math.sin(rx) + ((0 + 5 * math.sin(time * -2) + 10 * math.sin(time * 1 + 3)) * scale) * math.cos(rx)) * math.sin(ry)
    y = ((0) * scale) * math.cos(rx) - ((0 + 5 * math.sin(time * -2) + 10 * math.sin(time * 1 + 3)) * scale) * math.sin(rx)
    z = (-(0 + 5 * math.cos(time * -2) + 10 * math.cos(time * 1 + 3)) * scale) * math.sin(ry) + (((0) * scale) * math.sin(rx) + ((0 + 5 * math.sin(time * -2) + 10 * math.sin(time * 1 + 3)) * scale) * math.cos(rx)) * math.cos(ry)
    positions.append((x, y, z))
dictx = {}
dicty = {}

for i in range(-1,-181,-1):
    temp = -i
    dictx[i] = temp
for i in range(180,-1,-1):
    temp = 360-i
    dictx[i] = temp


file =  open("test.mcfunction","a",encoding="utf-8")
file.truncate(0)
# 输出粒子的位置列表
for i in positions:
    # print(f'particle dragonbreath ~{i[0]} ~2 ~{i[2]} {-i[0]} -1 {-i[2]} 0.2 0 normal')
    for x in range(-179,181):
        ceita = dictx[x]
        ceita_sin = math.sin(ceita_rd := math.radians(ceita))
        ceita_cos = math.cos(ceita_rd)

        prt = Rotate_3d.from_angle_axis(angle=90, axis=[0, 0, 1])
        result = prt.rotate_point(i)
        prt = Rotate_3d.from_angle_axis(angle=dictx[x]+90, axis=[0, 1, 0])
        result = prt.rotate_point(result)

        file.write(f"execute @s[rym={x},ry={x}] ~ ~ ~ particle dragonbreath ~{result[0]} ~{result[1] + 1.5} ~{result[2]} {1.5 * ceita_sin} 0 {1.5 * ceita_cos} 0.25 0 normal\n")
file.close()
