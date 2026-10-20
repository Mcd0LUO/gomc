# 方块状态表

1.12.2 的方块有属性，命令可以要**状态**而不是数据值：

```
/fill 2 4 2 3 4 3 stone variant=granite
/setblock 8 4 8 wool color=red
```

gomc 里这背后是两张表（`internal/version/v1_12_2/blockstates_table.go`）：

- **属性 → 取值 → 数据值**：每个方块有哪些属性、每个属性有哪些取值、每个
  取值组合对应哪个数据值；
- **数据值 → 数据值**：命令里写的数据值和 1.12.2 存下来的数据值**不是同一
  个数**（`setblock x y z chest 0` 放下去的是数据值 2 的箱子）。

两张表都是**跑出来的**，不是读字节码读出来的。

## 怎么重新生成

```powershell
$srv = 'D:\game\McServer\1.12.2'
$libs = (Get-ChildItem "$srv\libraries" -Recurse -Filter *.jar -File).FullName -join ';'
$cp = "$srv\cache\patched_1.12.2.jar;$libs"

# 1) 从真服务端 jar 里 dump
javac -encoding UTF-8 -cp $cp -d out DumpStates.java
java -cp "out;$cp" DumpStates > states.jsonl

# 2) 生成 Go 表
cd ../..
go run ./tools/blockstates -in tools/blockstates/states.jsonl `
    -out internal/version/v1_12_2/blockstates_table.go
```

`DumpStates.java` 会加载服务端的方块注册表（`DispenserRegistry.c()` 是
1.12.2 里那个 bootstrap 的入口），然后对每个方块问三件事：它有哪些属性、
每个属性有哪些取值、每个取值组合的数据值是多少；外加
`fromLegacyData(n)` 之后数据值是多少。

`states.jsonl` 是**已经 dump 好的**那份，进了版本库，所以没有服务端 jar
也能重新生成 Go 表。`internal/version/v1_12_2/blockstates_test.go` 里有一个
测试拿表和 dump 对账，表和 dump 对不上就会红。

## 为什么不用静态读字节码

属性是**运行期**构造的（`BlockStateEnum.of(name, EnumClass)` 从枚举常量
生成名字到取值的映射），静态读容易漏；而 `fromLegacyData` 的翻译关系是
各个方块自己覆写的，只有真跑一遍才准。

另外，Paper 的 patched jar 里的 NMS 类是 **SRG 重映射过**的
（`net.minecraft.server.v1_12_R1.BlockStateList` 这种名字），不用啃混淆名。

详细的流程见 [`../../docs/DEV.md`](../../docs/DEV.md)。
