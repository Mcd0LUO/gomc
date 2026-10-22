package v1_12_2

import (
	"github.com/Mcd0LUO/gomc/internal/nbt"
	"github.com/Mcd0LUO/gomc/internal/world"
)

// applyNBTTransform puts the tags that gomc models as entity fields back on
// the entity.
//
// 1.12.2 merges the tag of `/entitydata` into the entity's save compound and
// then calls `Entity.f(NBTTagCompound)` (readEntityFromNBT), which reads
// `Pos`, `Motion`, `Rotation`, ... back into the entity's fields. `/summon`
// goes through the same read: the entity factory builds the entity from the
// tag. gomc keeps an entity as its fields plus a tag compound, so the tags a
// command can observe have to be written to the fields as well.
//
// `Rotation` is the one that matters: it is what `@e[rx=...]` and
// `@e[ry=...]` select on. It is read the way `Entity.f` reads it --
// `getList("Rotation", 5)` keeps the list only when it is a **float** list,
// and a list that is not there reads as empty, where `NBTTagList.g(i)` is 0.
// So `Rotation:[90,45]` (ints) and `Rotation:[90.0d,45.0d]` (doubles) both
// put the entity at 0/0, while `Rotation:[90f,45f]` turns it. That is a
// 1.12.2 quirk, not a mistake.
//
// `Pose` is not applied. In 1.12.2 it is the limb angles of an armour stand
// (`EntityArmorStand`), it never reaches `yaw`/`pitch`, and gomc renders
// nothing, so it has no effect a command can observe; it stays a tag.
//
// `Pos` and `Motion` are not applied either. gomc does not simulate
// movement, and a teleport through a tag would bypass the chunk model the
// runtime keeps, so they stay tags as well.
func applyNBTTransform(e *world.Entity) {
	if e.NBT == nil {
		return
	}
	raw := e.NBT.Get("Rotation")
	if raw == nil {
		// A real entity always saves its rotation, so a tag compound
		// without one means nothing asked to change it: keep the
		// rotation the entity has.
		return
	}
	list, ok := raw.(*nbt.List)
	if !ok || list.ElemType() != nbt.TypeFloat {
		// `getList` rejects the tag, so the angles read as 0.
		e.Rot = world.Rotation{}
		return
	}
	e.Rot = world.Rotation{
		Yaw:   rotationAt(list, 0),
		Pitch: rotationAt(list, 1),
	}
}

// rotationAt reads one angle of a float list the way `NBTTagList.g(i)` does:
// a missing element is 0, and the value is a Java float, so it is rounded to
// 32 bits before it is kept.
func rotationAt(list *nbt.List, i int) float64 {
	f, ok := list.At(i).(nbt.Float)
	if !ok {
		return 0
	}
	return float64(f)
}

// syncRotationToNBT writes the rotation the entity has into its tag compound.
//
// 1.12.2 builds the compound `/entitydata` merges into from the entity
// (`CommandEntityData.a(entity)` saves it first), so the merge always starts
// from the rotation the entity currently has. gomc's tag compound can be
// older than the fields -- `/tp` turns the entity without touching the tags
// -- so the rotation is refreshed before the merge.
func syncRotationToNBT(e *world.Entity) {
	if e.NBT == nil {
		return
	}
	list := nbt.NewList(nbt.TypeFloat)
	list.Append(nbt.Float(e.Rot.Yaw))
	list.Append(nbt.Float(e.Rot.Pitch))
	e.NBT.Set("Rotation", list)
}
