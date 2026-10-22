# Relative teleports that land outside the bounds of `/tp`.
#
# The lines are parsed when the function is loaded, when the position the
# function will run at is not known, so the literal is inside the bounds
# and only the resolved coordinate is not: the first line is -1073741.824
# away from an entity that already stands near -30000000, and the second
# one drops the Y below -4096.
#
# The scoreboard tests read the `SuccessCount` channel that the failed
# command has just written, so a failure that reports nothing would show
# up as the old value.
tp @s ~-1073741.824 ~ ~
scoreboard players test @s t 0 0
tp @s ~ ~-5000 ~
scoreboard players test @s t 0 0
