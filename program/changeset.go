package program

type ChangeSet struct {
	BreakingChanges []string
	Commits         map[TypeTag][]string
}

func NewChangeset() *ChangeSet {
	return &ChangeSet{Commits: make(map[TypeTag][]string)}
}

func (c *ChangeSet) AddBreaking(message string) {
	c.BreakingChanges = append(c.BreakingChanges, message)
}

func (c *ChangeSet) AddCommit(tt TypeTag, section string, message string) {
	c.Commits[tt] = append(c.Commits[tt], message)
}
