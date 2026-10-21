scoreboard players add @s armor.putonStep 0


function rpg:hurt/player/armor/head_fail if @s[score_armor.test_min=6731,score_armor.test=6731,score_armor.putonStep_min=1]
function rpg:hurt/player/armor/head if @s[score_armor.test_min=6731,score_armor.test=6731,score_armor.putonStep=0]


function rpg:hurt/player/armor/chest_fail_2 if @s[score_armor.test_min=6732,score_armor.test=6732,score_armor.putonStep=0]
function rpg:hurt/player/armor/chest_fail if @s[score_armor.test_min=6732,score_armor.test=6732,score_armor.putonStep_min=2]
function rpg:hurt/player/armor/chest if @s[score_armor.test_min=6732,score_armor.test=6732,,score_armor.putonStep_min=1,score_armor.putonStep=1]

function rpg:hurt/player/armor/leggings_fail_2 if @s[score_armor.test_min=6733,score_armor.test=6733,score_armor.putonStep=1]
function rpg:hurt/player/armor/leggings_fail if @s[score_armor.test_min=6733,score_armor.test=6733,score_armor.putonStep_min=3]
function rpg:hurt/player/armor/leggings if @s[score_armor.test_min=6733,score_armor.test=6733,score_armor.putonStep_min=2,score_armor.putonStep=2]


function rpg:hurt/player/armor/boots_fail if @s[score_armor.test_min=6734,score_armor.test=6734,score_armor.putonStep=2]
function rpg:hurt/player/armor/boots if @s[score_armor.test_min=6734,score_armor.test=6734,score_armor.putonStep_min=3]

advancement revoke @s from main:adv/rpg/main