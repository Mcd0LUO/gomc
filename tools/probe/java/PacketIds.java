import net.minecraft.server.v1_12_R1.EnumProtocol;
import net.minecraft.server.v1_12_R1.EnumProtocolDirection;
import net.minecraft.server.v1_12_R1.PacketPlayInKeepAlive;
import net.minecraft.server.v1_12_R1.PacketPlayOutKeepAlive;

/** Prints the wire ids of the keep alive packets of 1.12.2. */
public class PacketIds {
    public static void main(String[] args) throws Exception {
        System.out.println("clientbound keep alive = 0x" + Integer.toHexString(
                EnumProtocol.PLAY.a(EnumProtocolDirection.CLIENTBOUND, new PacketPlayOutKeepAlive(0L))));
        System.out.println("serverbound keep alive = 0x" + Integer.toHexString(
                EnumProtocol.PLAY.a(EnumProtocolDirection.SERVERBOUND, new PacketPlayInKeepAlive())));
    }
}
