import math

scale = 0.5
rx = 0  # x轴旋转角度，单位为弧度
rz = 45  # z轴旋转角度，单位为弧度
vy = 0
time_step = 0.1
total_time = 60

for t in range(0, int(total_time / time_step)):
    t_actual = t * time_step
    x = ((0 + 6 * math.cos(t_actual * -1 + 0) + 2 * math.cos(t_actual * 3 + 0)) * scale) * math.cos(rz) - (((0) * scale) * math.cos(rx) - ((0 + 6 * math.sin(t_actual * -1 + 0) + 2 * math.sin(t_actual * 3 + 0)) * scale) * math.sin(rx)) * math.sin(rz)
    y = ((0 + 6 * math.cos(t_actual * -1 + 0) + 2 * math.cos(t_actual * 3 + 0)) * scale) * math.sin(rz) + (((0) * scale) * math.cos(rx) - ((0 + 6 * math.sin(t_actual * -1 + 0) + 2 * math.sin(t_actual * 3 + 0)) * scale) * math.sin(rx)) * math.cos(rz)
    z = ((0) * scale) * math.sin(rx) + ((0 + 6 * math.sin(t_actual * -1 + 0) + 2 * math.sin(t_actual * 3 + 0)) * scale) * math.cos(rx)

    print(f"particle endRod ~{x} ~{y} ~{z} 0 0 0 0.2 0")
