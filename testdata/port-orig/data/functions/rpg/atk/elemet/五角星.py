import math

scale = 0.3
rx = 0.5236
ry = -0.1745

delta_t = 0.025
total_time = 60

x = 0
y = 0
z = 0

for t in range(0, int(total_time/delta_t)):
    x = ((0 + 15 * math.cos(t * -2 + 0) + 8 * math.cos(t * 3 + 3)) * scale) * math.cos(ry) + (((0) * scale) * math.sin(rx) + ((0 + 15 * math.sin(t * -2 + 0) + 8 * math.sin(t * 3 + 3)) * scale) * math.cos(rx)) * math.sin(ry)
    y = ((0) * scale) * math.cos(rx) - ((0 + 15 * math.sin(t * -2 + 0) + 8 * math.sin(t * 3 + 3)) * scale) * math.sin(rx)
    z = (-(0 + 15 * math.cos(t * -2 + 0) + 8 * math.cos(t * 3 + 3)) * scale) * math.sin(ry) + (((0) * scale) * math.sin(rx) + ((0 + 15 * math.sin(t * -2 + 0) + 8 * math.sin(t * 3 + 3)) * scale) * math.cos(rx)) * math.cos(ry)


    # 输出粒子的位置
    print(f" {x}, {y}, {z})")
