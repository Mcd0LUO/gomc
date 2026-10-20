import com.google.common.base.Optional;

import net.minecraft.server.v1_12_R1.Block;
import net.minecraft.server.v1_12_R1.BlockStateList;
import net.minecraft.server.v1_12_R1.IBlockData;
import net.minecraft.server.v1_12_R1.IBlockState;
import net.minecraft.server.v1_12_R1.INamable;
import net.minecraft.server.v1_12_R1.MinecraftKey;
import net.minecraft.server.v1_12_R1.RegistryBlocks;

import java.io.PrintStream;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.List;

/**
 * 把 1.12.2 服务端 jar 里的「方块 -> 状态属性」表 dump 成 JSON。
 *
 * 为什么用运行时而不是读字节码：
 * 1.12.2 的属性系统是运行期构造的（PropertyEnum.of(name, EnumClass) 会
 * 从枚举常量生成 name->value 映射），静态读容易漏。
 * 2. 这里跑的就是真服用的那份 jar（patched_1.12.2.jar），出来的表就是
 *    真服命令认的那张表。
 *
 * 输出格式（每行一个 JSON 块，方便 grep / 逐行 diff）：
 *   {"block":"minecraft:stone","kind":"enum","props":[...],"states":[...]}
 */
public class DumpStates {

    public static void main(String[] args) throws Exception {
        PrintStream out = new PrintStream(System.out, true, "UTF-8");

        // Blocks 的静态初始化块会检查 "Bootstrap 跑过没有"，没跑直接抛
        // "Accessed Blocks before Bootstrap!"。DispenserRegistry.c() 就是
        // 1.12.2 里那个 bootstrap 的入口。
        try {
            net.minecraft.server.v1_12_R1.DispenserRegistry.c();
        } catch (Throwable t) {
            System.err.println("bootstrap 失败：" + t);
            t.printStackTrace();
        }
        try {
            Class.forName("net.minecraft.server.v1_12_R1.Blocks");
        } catch (Throwable t) {
            System.err.println("初始化 Blocks 失败：");
            Throwable c = t;
            while (c != null) {
                System.err.println("  " + c.getClass().getName() + ": " + c.getMessage());
                for (StackTraceElement e : c.getStackTrace()) {
                    System.err.println("      at " + e);
                }
                c = c.getCause();
            }
            System.exit(2);
        }

        RegistryBlocks<MinecraftKey, Block> registry = Block.REGISTRY;

        for (MinecraftKey key : registry.keySet()) {
            Block block = registry.get(key);
            if (block == null) {
                continue;
            }
            BlockStateList list = block.s();
            Collection<IBlockState<?>> props = list == null
                    ? Collections.<IBlockState<?>>emptyList() : list.d();
            if (props == null) {
                props = Collections.<IBlockState<?>>emptyList();
            }

            // 属性清单：名字 + 取值。取值用的是「命令里能写的那串」：
            // 枚举走 INamable.getName()，整数就是数字本身。
            // 种类也要记下来：整数属性是按范围解析的（"007" 在真服上是 7），
            // 枚举和布尔是按名字精确匹配的。
            StringBuilder propJson = new StringBuilder();
            propJson.append('[');
            boolean first = true;
            for (IBlockState<?> prop : props) {
                if (!first) {
                    propJson.append(',');
                }
                first = false;
                propJson.append("{\"name\":").append(quote(prop.a()))
                        .append(",\"kind\":").append(quote(kindOf(prop)))
                        .append(",\"values\":[");
                Collection<?> values = prop.c();
                boolean firstVal = true;
                if (values != null) {
                    for (Object v : values) {
                        if (!firstVal) {
                            propJson.append(',');
                        }
                        firstVal = false;
                        propJson.append(quote(valueName(prop, v)));
                    }
                }
                propJson.append("]}");
            }
            propJson.append(']');

            // 全部状态：各属性取值的**完整笛卡尔积**。命令里 "k=v" 是打在
            // 默认状态上的，所以要能算出「只有部分属性被改」时对应的
            // legacy 数据值。
            //
            // 展开顺序是「第一个属性是最外层」，也就是行主序：
            //   下标 = ((i0 * n1) + i1) * n2 + i2 ...
            // 生成 Go 表时只要按这个顺序存 meta 数组就够了。
            IBlockData base = block.getBlockData();
            List<IBlockData> all = new ArrayList<IBlockData>();
            enumerate(block, base, props, 0, all);

            // 默认值：每个属性在默认状态里是第几个取值。
            StringBuilder defJson = new StringBuilder();
            defJson.append('[');
            boolean firstDef = true;
            for (IBlockState<?> prop : props) {
                if (!firstDef) {
                    defJson.append(',');
                }
                firstDef = false;
                int idx = 0;
                Object dv = getProp(base, prop);
                for (Object v : prop.c()) {
                    if (valueName(prop, v).equals(valueName(prop, dv))) {
                        break;
                    }
                    idx++;
                }
                defJson.append(idx);
            }
            defJson.append(']');

            StringBuilder metaJson = new StringBuilder();
            metaJson.append('[');
            boolean firstState = true;
            for (IBlockData data : all) {
                if (!firstState) {
                    metaJson.append(',');
                }
                firstState = false;
                metaJson.append(block.toLegacyData(data));
            }
            metaJson.append(']');

            out.println("{\"block\":" + quote(key.toString())
                    + ",\"props\":" + propJson
                    + ",\"default\":" + defJson
                    + ",\"defaultMeta\":" + block.toLegacyData(base)
                    + ",\"metas\":" + metaJson
                    + ",\"legacy\":" + legacyJson(block) + "}");
        }
    }

    /**
     * 数据值的翻译表：命令里写的数据值 n，放下去的状态的数据值是
     * toLegacyData(fromLegacyData(n))。
     *
     * 对大多数方块这是恒等映射，但有例外：箱子写 0 放下去的是朝北的箱子
     * （数据值 2），没有属性的方块写 5 放下去的也是数据值 0。
     */
    private static String legacyJson(Block block) {
        StringBuilder sb = new StringBuilder();
        sb.append('[');
        for (int n = 0; n < 16; n++) {
            if (n > 0) {
                sb.append(',');
            }
            int meta;
            try {
                meta = block.toLegacyData(block.fromLegacyData(n));
            } catch (Throwable t) {
                meta = -1;
            }
            sb.append(meta);
        }
        return sb.append(']').toString();
    }

    /** 深度优先地展开所有属性取值的组合。 */
    private static void enumerate(Block block, IBlockData data,
                                  Collection<IBlockState<?>> props, int idx,
                                  List<IBlockData> out) {
        if (idx == props.size()) {
            out.add(data);
            return;
        }
        IBlockState<?> prop = null;
        int i = 0;
        for (IBlockState<?> p : props) {
            if (i++ == idx) {
                prop = p;
                break;
            }
        }
        Collection<?> values = prop.c();
        if (values == null) {
            enumerate(block, data, props, idx + 1, out);
            return;
        }
        for (Object v : values) {
            IBlockData next = setProp(data, prop, v);
            enumerate(block, next, props, idx + 1, out);
        }
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    private static IBlockData setProp(IBlockData data, IBlockState prop, Object value) {
        return ((IBlockData) data.set((IBlockState) prop, (Comparable) value));
    }

    @SuppressWarnings({"unchecked", "rawtypes"})
    private static Object getProp(IBlockData data, IBlockState prop) {
        return data.get((IBlockState) prop);
    }

    /** 属性的种类：integer 的取值是范围，别的按名字精确匹配。 */
    private static String kindOf(IBlockState<?> prop) {
        if (prop instanceof net.minecraft.server.v1_12_R1.BlockStateInteger) {
            return "integer";
        }
        if (prop instanceof net.minecraft.server.v1_12_R1.BlockStateBoolean) {
            return "boolean";
        }
        if (prop instanceof net.minecraft.server.v1_12_R1.BlockStateDirection) {
            return "direction";
        }
        return "enum";
    }

    /** 取「命令里能写的那个串」：枚举用 INamable.getName()，其余用 toString。 */
    private static String valueName(IBlockState<?> prop, Object value) {
        if (value instanceof INamable) {
            return ((INamable) value).getName();
        }
        return String.valueOf(value);
    }

    private static String quote(String s) {
        StringBuilder sb = new StringBuilder("\"");
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"':
                    sb.append("\\\"");
                    break;
                case '\\':
                    sb.append("\\\\");
                    break;
                case '\n':
                    sb.append("\\n");
                    break;
                case '\r':
                    sb.append("\\r");
                    break;
                case '\t':
                    sb.append("\\t");
                    break;
                default:
                    if (c < 0x20) {
                        sb.append(String.format("\\u%04x", (int) c));
                    } else {
                        sb.append(c);
                    }
            }
        }
        return sb.append('"').toString();
    }
}
