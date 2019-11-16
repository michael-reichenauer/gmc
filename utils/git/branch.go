package git

type Branch struct {
	ID               string
	Name             string
	TipID            string
	IsCurrent        bool
	IsRemote         bool
	RemoteName       string
	IsDetached       bool
	AheadCount       int
	BehindCount      int
	IsRemoteMissing  bool
	TipCommitMessage string
}

func (b *Branch) String() string {
	return b.ID
}
