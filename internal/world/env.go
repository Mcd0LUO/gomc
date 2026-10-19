package world

// World state that the environment commands read and write.
//
// 1.12.2 keeps this state in the world and the server, and the commands
// `/weather`, `/difficulty`, `/setworldspawn`, `/spawnpoint`,
// `/worldborder` and `/setidletimeout` are the only ways to change it.

// WeatherKind is a weather state.
type WeatherKind string

const (
	// WeatherClear is the clear weather.
	WeatherClear WeatherKind = "clear"
	// WeatherRain is the rain.
	WeatherRain WeatherKind = "rain"
	// WeatherThunder is the thunderstorm.
	WeatherThunder WeatherKind = "thunder"
)

// WeatherState is the world's weather.
type WeatherState struct {
	// Kind is the current weather.
	Kind WeatherKind
	// Duration is the remaining duration in ticks, 0 for an indefinite
	// rain.
	Duration int
	// ClearDuration is the remaining number of ticks of clear weather.
	ClearDuration int
	// Storm reports whether it is raining.
	Storm bool
	// Thundering reports whether there is a thunderstorm.
	Thundering bool
}

// Clear reports whether the weather is clear.
func (w WeatherState) Clear() bool { return w.Kind == WeatherClear }

// Difficulty is a world difficulty.
type Difficulty string

const (
	// DifficultyPeaceful is the peaceful difficulty.
	DifficultyPeaceful Difficulty = "peaceful"
	// DifficultyEasy is the easy difficulty.
	DifficultyEasy Difficulty = "easy"
	// DifficultyNormal is the normal difficulty.
	DifficultyNormal Difficulty = "normal"
	// DifficultyHard is the hard difficulty.
	DifficultyHard Difficulty = "hard"
)

// DifficultyIDs maps a difficulty to its 1.12.2 id.
//
// The ids are the ones `EnumDifficulty` uses: 0 peaceful, 1 easy,
// 2 normal, 3 hard.
var DifficultyIDs = map[Difficulty]int{
	DifficultyPeaceful: 0,
	DifficultyEasy:     1,
	DifficultyNormal:   2,
	DifficultyHard:     3,
}

// ValidDifficulty reports whether a name is a known difficulty.
func ValidDifficulty(name string) bool {
	_, ok := DifficultyIDs[Difficulty(name)]
	return ok
}

// WorldBorder is the world border.
type WorldBorder struct {
	// Center is the border's center.
	Center Vec
	// Size is the border's diameter, in blocks.
	Size float64
	// WarningDistance is how far inside the border the warning appears.
	WarningDistance float64
	// WarningTime is the warning delay in seconds.
	WarningTime int
	// DamagePerBlock is the damage per block travelled inside.
	DamagePerBlock float64
	// LerpSize is the size the border shrinks towards, in blocks.
	LerpSize float64
}

// DefaultWorldBorder returns the border a fresh world has.
//
// The defaults are the ones 1.12.2 starts with: a border of 60 million
// blocks, a warning at 5 blocks and 15 seconds.
func DefaultWorldBorder() WorldBorder {
	return WorldBorder{
		Center:          Vec{},
		Size:            60000000,
		WarningDistance: 5,
		WarningTime:     15,
		DamagePerBlock:  0.2,
	}
}
