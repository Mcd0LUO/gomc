# 弹出主类
# 分选
function sys:api/文显/eject/call-1-50 if @s[score_return_min=1,score_return=50]
function sys:api/文显/eject/call-51-100 if @s[score_return_min=51,score_return=100]
function sys:api/文显/eject/call-101-150 if @s[score_return_min=101,score_return=150]
function sys:api/文显/eject/call-151-200 if @s[score_return_min=151,score_return=200]
# 移除标签
scoreboard players tag @s remove new