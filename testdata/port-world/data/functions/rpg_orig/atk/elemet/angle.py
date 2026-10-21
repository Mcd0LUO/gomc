from math import sin, cos, radians
'''
基于球体参数方程
'''
dictx = {}
dicty = {}
r=  0.5
sr = 0.8
cmd = ""
# dictx 的键是游戏内z轴，值是转换后的x轴坐标
for i in range(-1,-181,-1):
    temp = -i
    dictx[i] = temp
for i in range(180,-1,-1):
    temp = 360-i
    dictx[i] = temp
# dicty 的键是游戏内y轴,值是转换后的z轴坐标
for i in range(-90,91):
    temp = i+90
    dicty[i] = temp

file = open("test.mcfunction", 'a', encoding='utf-8')
file.truncate(0)

for x in range(-180,180,30):
    ceita = dictx[x]
    ceita_sin = sin(ceita_rd:=radians(ceita))
    ceita_cos = cos(ceita_rd)
    for y in range(0,1):
        fai = dicty[y]
        fai_sin = sin(fai_rd := radians(fai))
        fai_cos = cos(fai_rd)
        # 球
        #cmd = f'particle flame ~{round(3 * fai_sin * ceita_sin, 6)} ~{round(3 * fai_cos,6)} ~{round(3 * fai_sin * ceita_cos, 6)} {-round(2 * fai_sin * ceita_sin, 6)} {-round(0.5 * fai_cos,6)} {-round(2 * fai_sin * ceita_cos, 6)} 0.25 0\n'
        # 放大圆
        cmd = f'particle bubble ~{round(1.5 * fai_sin * ceita_sin, 6)} ~1 ~{round(1.5 * fai_sin * ceita_cos, 6)} 0 0.5 0 0.4 10 normal\n'
        print(cmd,end='')
        #cmd = cmd.replace("[**", "{").replace("**]", "}")
        #file.write(cmd)
file.close