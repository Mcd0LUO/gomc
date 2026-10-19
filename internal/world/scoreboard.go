package world

import (
	"fmt"
	"sort"
)

// Criteria is a scoreboard objective criterion.
type Criteria string

// The criteria supported by 1.12.2 scoreboard objectives.
const (
	// CriteriaDummy is the writable-by-command-only criterion.
	CriteriaDummy Criteria = "dummy"
	// CriteriaTrigger is writable through /trigger.
	CriteriaTrigger Criteria = "trigger"
	// CriteriaPlayerKill counts player kills.
	CriteriaPlayerKill Criteria = "playerKillCount"
	// CriteriaDeath counts deaths.
	CriteriaDeath Criteria = "deathCount"
	// CriteriaTotalKill counts all kills.
	CriteriaTotalKill Criteria = "totalKillCount"
	// CriteriaHealth tracks player health.
	CriteriaHealth Criteria = "health"
	// CriteriaXP tracks experience.
	CriteriaXP Criteria = "xp"
	// CriteriaLevel tracks experience level.
	CriteriaLevel Criteria = "level"
	// CriteriaFood tracks saturation.
	CriteriaFood Criteria = "food"
	// CriteriaAir tracks air supply.
	CriteriaAir Criteria = "air"
	// CriteriaArmor tracks armor.
	CriteriaArmor Criteria = "armor"
)

// ReadOnly reports whether the criterion cannot be modified by commands.
func (c Criteria) ReadOnly() bool {
	switch c {
	case CriteriaHealth, CriteriaXP, CriteriaLevel, CriteriaFood, CriteriaAir, CriteriaArmor:
		return true
	}
	return false
}

// Valid reports whether the criterion is known.
func (c Criteria) Valid() bool {
	switch c {
	case CriteriaDummy, CriteriaTrigger, CriteriaPlayerKill, CriteriaDeath,
		CriteriaTotalKill, CriteriaHealth, CriteriaXP, CriteriaLevel,
		CriteriaFood, CriteriaAir, CriteriaArmor:
		return true
	}
	return false
}

// Objective is a scoreboard objective.
type Objective struct {
	// Name is the objective's identifier.
	Name string
	// Criteria is the objective's criterion.
	Criteria Criteria
	// DisplayName is the optional display name.
	DisplayName string
}

// Scoreboard holds objectives, player scores and teams.
type Scoreboard struct {
	objectives map[string]*Objective
	scores     map[string]map[string]int
	teams      map[string]*Team
	// order preserves objective creation order for stable listing.
	order []string
}

// NewScoreboard returns an empty scoreboard.
func NewScoreboard() *Scoreboard {
	return &Scoreboard{
		objectives: map[string]*Objective{},
		scores:     map[string]map[string]int{},
		teams:      map[string]*Team{},
	}
}

// AddObjective creates an objective, reporting whether it was newly created.
func (s *Scoreboard) AddObjective(name string, criteria Criteria, display string) (*Objective, bool) {
	if o, ok := s.objectives[name]; ok {
		return o, false
	}
	o := &Objective{Name: name, Criteria: criteria, DisplayName: display}
	s.objectives[name] = o
	s.order = append(s.order, name)
	return o, true
}

// RemoveObjective deletes an objective and all its scores.
func (s *Scoreboard) RemoveObjective(name string) bool {
	if _, ok := s.objectives[name]; !ok {
		return false
	}
	delete(s.objectives, name)
	delete(s.scores, name)
	for i, n := range s.order {
		if n == name {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

// Objective returns an objective by name.
func (s *Scoreboard) Objective(name string) *Objective { return s.objectives[name] }

// Objectives returns the objectives in creation order.
func (s *Scoreboard) Objectives() []*Objective {
	out := make([]*Objective, 0, len(s.order))
	for _, n := range s.order {
		out = append(out, s.objectives[n])
	}
	return out
}

// SetScore stores a score, reporting whether the holder was newly tracked.
func (s *Scoreboard) SetScore(holder, objective string, value int) bool {
	if _, ok := s.scores[objective]; !ok {
		s.scores[objective] = map[string]int{}
	}
	_, existed := s.scores[objective][holder]
	s.scores[objective][holder] = wrapInt32(value)
	return !existed
}

// Score returns a holder's score and whether it is tracked.
func (s *Scoreboard) Score(holder, objective string) (int, bool) {
	m, ok := s.scores[objective]
	if !ok {
		return 0, false
	}
	v, ok := m[holder]
	return v, ok
}

// ResetScore removes a holder's score from an objective. Passing an empty
// objective removes the holder from every objective.
func (s *Scoreboard) ResetScore(holder, objective string) {
	if objective == "" {
		for _, m := range s.scores {
			delete(m, holder)
		}
		return
	}
	if m, ok := s.scores[objective]; ok {
		delete(m, holder)
	}
}

// Holders returns the holders tracked by an objective, sorted.
func (s *Scoreboard) Holders(objective string) []string {
	m := s.scores[objective]
	out := make([]string, 0, len(m))
	for h := range m {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

// Operation is a scoreboard arithmetic operation.
type Operation string

// The scoreboard operations supported by 1.12.2.
//
// `<` and `>` keep the smaller/larger value respectively, and `><` swaps
// the two values.
const (
	// OpAdd adds the operand to the target.
	OpAdd Operation = "+="
	// OpSub subtracts the operand from the target.
	OpSub Operation = "-="
	// OpMul multiplies the target by the operand.
	OpMul Operation = "*="
	// OpDiv divides the target by the operand.
	OpDiv Operation = "/="
	// OpMod takes the remainder of the target divided by the operand.
	OpMod Operation = "%="
	// OpSet replaces the target with the operand.
	OpSet Operation = "="
	// OpKeepMin keeps the smaller of the two values.
	OpKeepMin Operation = "<"
	// OpKeepMax keeps the larger of the two values.
	OpKeepMax Operation = ">"
	// OpSwap exchanges the two values.
	OpSwap Operation = "><"
)

// Valid reports whether the operation is known.
func (o Operation) Valid() bool {
	switch o {
	case OpAdd, OpSub, OpMul, OpDiv, OpMod, OpSet, OpKeepMin, OpKeepMax, OpSwap:
		return true
	}
	return false
}

// Apply computes the result of applying op to the target and operand values.
//
// The arithmetic follows Java's integer semantics: division truncates toward
// zero, `%` takes the sign of the dividend, and every result wraps to a
// signed 32-bit integer the way a Java `int` does.
func Apply(op Operation, target, operand int) (int, error) {
	switch op {
	case OpAdd:
		return wrapInt32(target + operand), nil
	case OpSub:
		return wrapInt32(target - operand), nil
	case OpMul:
		return wrapInt32(target * operand), nil
	case OpDiv:
		if operand == 0 {
			return 0, fmt.Errorf("cannot divide by zero")
		}
		return wrapInt32(target / operand), nil
	case OpMod:
		if operand == 0 {
			return 0, fmt.Errorf("cannot divide by zero")
		}
		return wrapInt32(target % operand), nil
	case OpSet:
		return wrapInt32(operand), nil
	case OpKeepMin:
		if operand < target {
			return wrapInt32(operand), nil
		}
		return wrapInt32(target), nil
	case OpKeepMax:
		if operand > target {
			return wrapInt32(operand), nil
		}
		return wrapInt32(target), nil
	case OpSwap:
		return wrapInt32(operand), nil
	}
	return 0, fmt.Errorf("unknown operation %q", op)
}

// wrapInt32 folds a value into the signed 32-bit range.
//
// 1.12.2 stores every score in a Java `int`, so the arithmetic wraps around
// instead of growing without bound.
func wrapInt32(v int) int {
	return int(int32(v))
}

// Team is a scoreboard team.
type Team struct {
	// Name is the team's identifier.
	Name string
	// DisplayName is the optional display name.
	DisplayName string
	// Members are the team members.
	Members map[string]bool
	// Options holds the team options such as color and collision rules.
	Options map[string]string
}

// NewTeam returns an empty team.
func NewTeam(name string) *Team {
	return &Team{Name: name, Members: map[string]bool{}, Options: map[string]string{}}
}

// AddTeam registers a team, reporting whether it was newly created.
func (s *Scoreboard) AddTeam(name string) (*Team, bool) {
	if t, ok := s.teams[name]; ok {
		return t, false
	}
	t := NewTeam(name)
	s.teams[name] = t
	return t, true
}

// RemoveTeam deletes a team.
func (s *Scoreboard) RemoveTeam(name string) bool {
	if _, ok := s.teams[name]; !ok {
		return false
	}
	delete(s.teams, name)
	return true
}

// Team returns a team by name.
func (s *Scoreboard) Team(name string) *Team { return s.teams[name] }

// Teams returns all teams, sorted by name.
func (s *Scoreboard) Teams() []*Team {
	out := make([]*Team, 0, len(s.teams))
	for _, t := range s.teams {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// TeamOf returns the team a holder belongs to, or nil.
func (s *Scoreboard) TeamOf(holder string) *Team {
	for _, t := range s.teams {
		if t.Members[holder] {
			return t
		}
	}
	return nil
}

// TeamNameOf returns the name of the team a holder belongs to.
//
// An entity that is on no team reports an empty name, which is what
// 1.12.2's `team=` filter compares against.
func (s *Scoreboard) TeamNameOf(holder string) string {
	if t := s.TeamOf(holder); t != nil {
		return t.Name
	}
	return ""
}

// AddTeamMember adds a holder to a team.
//
// The holder leaves any other team first, because 1.12.2 keeps a holder in
// at most one team.
func (s *Scoreboard) AddTeamMember(team, holder string) bool {
	t := s.teams[team]
	if t == nil {
		return false
	}
	for _, other := range s.teams {
		delete(other.Members, holder)
	}
	t.Members[holder] = true
	return true
}

// RemoveTeamMember removes a holder from a team.
func (s *Scoreboard) RemoveTeamMember(team, holder string) bool {
	t := s.teams[team]
	if t == nil {
		return false
	}
	delete(t.Members, holder)
	return true
}
