# 生成契约兽
tellraw @s ["┍─────────────当前契约兽─────────────┑"]

tellraw @s[score_pet.now_min=1,score_pet.now=1] ["§r<§9§l☁§r> §e⚔ §r银月§b啸嗷             ",{"text":"§r<§a❀ §e召唤 §a❀§r>","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 5101"}}]


tellraw @s[score_pet.now_min=1] [{"text":"\n§r<§6❂ §e战兽召回 §6❂§r>","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 5007"},"hoverEvent":{"action":"show_text","value":"§7将§e已召唤§7的§e契约战兽§7传送至您§b身边"}},{"text":"          "},{"text":"§r<§6❂ §e战兽入库 §6❂§r>","clickEvent":{"action":"run_command","value":"/trigger menu.trigger set 5008"},"hoverEvent":{"action":"show_text","value":"§7将§e已召唤§7的§e契约兽§7收回§b契约目录"}}]