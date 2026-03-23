package repo

// ListOpts filters repository listing for any backend.
type ListOpts struct {
	Mine       bool
	Org        string
	NoArchived bool
	OnlyForks  bool
	NoForks    bool
}
