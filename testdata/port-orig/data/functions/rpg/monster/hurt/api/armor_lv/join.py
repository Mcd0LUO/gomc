import os
try:
    minn = 1
    maxn =5

    while maxn<=100:
       # os.makedirs(f"{minn}-{maxn}")
        with open(f"{minn}-{maxn}.mcfunction",'a+') as f:
            for i in range(minn,maxn+1):
              f.write(f"scoreboard teams join {i} @s[score_return_min={i},score_return={i}]")
              if i%5 != 0:
                  f.write("\n")
        minn,maxn = minn+5,maxn+5
finally:
    pass