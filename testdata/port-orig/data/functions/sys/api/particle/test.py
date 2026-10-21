from math import sin, cos, radians
from random import randint
'''
基于球体参数方程
'''
dictx = {}
dicty = {}
r=  1
sr = 0.8
cmd = ""
time = 0
for i in range(-1,-181,-1):
    temp = -i
    dictx[i] = temp
for i in range(180,-1,-1):
    temp = 360-i
    dictx[i] = temp
for i in range(-90,91):
    temp = i+90
    dicty[i] = temp

file = open("test.mcfunction", 'a', encoding='utf-8')
file.truncate(0)

for x in range(-180,180,40):
    ceita = dictx[x]
    ceita_sin = sin(ceita_rd:=radians(ceita))
    ceita_cos = cos(ceita_rd)
    for y in range(0,1):
        fai = dicty[y]
        fai_sin = sin(fai_rd := radians(fai))
        fai_cos = cos(fai_rd)
        cmd = f'particle flame ~{round(3 * fai_sin * ceita_sin, 6)} ~{round(3 * fai_cos,6)} ~{round(3 * fai_sin * ceita_cos, 6)} {-round(2 * fai_sin * ceita_sin, 6)} {-round(0.5 * fai_cos,6)} {-round(2 * fai_sin * ceita_cos, 6)} 0.25 0\n'

file.close()
