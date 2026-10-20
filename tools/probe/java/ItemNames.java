import net.minecraft.server.v1_12_R1.Item;
import net.minecraft.server.v1_12_R1.ItemStack;
import net.minecraft.server.v1_12_R1.LocaleI18n;
import net.minecraft.server.v1_12_R1.MinecraftKey;

import java.util.Map;

/**
 * Prints the display name of every item of 1.12.2 for every data value.
 *
 * The names come from the server itself: the item registry of the server jar
 * is walked and the name of the stack is asked for, so the table is the one
 * the server prints, not a guess.
 *
 * Usage: java -cp <server jar>:<libraries> ItemNames
 */
public class ItemNames {
    public static void main(String[] args) {
        // The server registers the blocks and the items when it starts.
        // "DispenserRegistry.c()" is the obfuscated "Bootstrap.register()".
        net.minecraft.server.v1_12_R1.DispenserRegistry.c();
        System.out.println("# key\thasSubtypes\tdata0\tdata1\t... data15");
        for (MinecraftKey key : Item.REGISTRY.keySet()) {
            Item item = Item.REGISTRY.get(key);
            StringBuilder sb = new StringBuilder();
            sb.append(key.getKey()).append('\t').append(item.k());
            boolean varies = false;
            String first = null;
            for (int data = 0; data < 16; data++) {
                String name = new ItemStack(item, 1, data).getName();
                if (first == null) {
                    first = name;
                } else if (!first.equals(name)) {
                    varies = true;
                }
                sb.append('\t').append(name);
            }
            sb.append('\t').append(varies ? "VARIES" : "fixed");
            System.out.println(sb);
        }
        // A sanity check: the language file has to be on the class path.
        System.out.println("# sanity: " + LocaleI18n.get("item.dyePowder.black.name"));
    }
}
