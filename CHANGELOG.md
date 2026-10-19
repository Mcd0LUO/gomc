# 更新日志

本文件记录 gomc 的重要变化。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循[语义化版本](https://semver.org/lang/zh-CN/)。

0.x.y 表示这个版本的行为**还在对齐中**：0.1.0 已经有完整的命令面，
但细节仍可能与真实服务端有出入（已知偏差见 `docs/SPEC-1.12.2.md` 的"还没对齐的"一节）。
1.0.0 会等到 1.12.2 的行为被完整复现。

## [未发布]

### 计划中

- 选择器语义收尾：体积盒、整数半径、负半径、`score_x` 与 `score_x_min` 并存。
- 命令回显逐条对齐语言文件（含 `scoreboard teams`、`/give` 这些目前缺回显的）。
- 方块注册表：让写错的方块名像真服一样报错。

## [0.1.0] - 2026-10-20

首个可用版本。

### 新增

- **1.12.2 命令解释器**：实现 1.12.2 专用服务器那 64 条命令中的 46 条，覆盖写 mcfunction
  最常用的那些（`execute`、`scoreboard`、`summon`、`setblock`、`fill`、`clone`、
  `function`、`gamerule`、`time`、`weather`……）。
- **目标选择器**：`@p` / `@r` / `@a` / `@e` / `@s` 与 1.12.2 的全部过滤器
  （距离、旋转、体积、分数、游戏模式、标签、队伍、名称、类型）。
- **对拍层**（`internal/difftest`）：用 YAML 风格用例描述一段场景，把执行结果与
  golden 文件比对，锁住行为；`difftest probe` 还能生成一份能丢到真实服务端上跑的
  探针函数。
- **真服探针工具**（`tools/probe`）：在隔离的 1.12.2 服务端上跑用例、抓真实输出，
  用来验证规格。隔离环境用 junction 复用依赖，用户自己的存档与进程不受影响。
- **增强报错**（默认开启）：报出文件、行、列、那一行原文、插入符、正确用法和
  `did you mean` 建议；`--plain-errors` 可关掉增强只留一行。
  错误**原文**始终是 1.12.2 的原文，增强只是附加。
- **语义化版本号**：`gomc --version` 打印版本、提交号与构建时间。

### 行为对齐

真服实测（不是查 Wiki）定下来的几条关键行为：

- `execute` 对**每个**匹配实体各执行一次，顺序是**距离升序**（不是倒序）。
- `execute` 里一个实体失败**不影响**其余实体。
- 实体的显示名是**翻译名**（`Armor Stand`），不是注册名。
- 1.12.2 **没有**独立的 `/tag` 命令，标签走 `scoreboard players tag`。
- 标签就是实体 NBT 里的 `Tags`，写 NBT 和用计分板命令是同一份数据。
- `scoreboard players test` 只给一个值时，上限是 `2147483647`（不是等于下限）。
- 函数 id 按 `MinecraftKey` 小写化，缺命名空间默认 `minecraft:`。

### 已知边界

- 只做 1.12.2；1.13+ 的 `execute as/at/run`、`scores={...}`、datapack 布局都不支持。
- 不模拟实体 AI、寻路、重力、区块生成、渲染和网络协议。
- 方块名不校验（写错方块名不会报错）。
- 实体没有建模 UUID，计分板持有者用内部 id。

[未发布]: https://github.com/mcsim/gomc/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mcsim/gomc/releases/tag/v0.1.0
