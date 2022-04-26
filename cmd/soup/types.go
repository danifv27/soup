package main

type contextStr string

type WasSetted struct {
	contextWasSet bool
}

func (f contextStr) BeforeApply(set *WasSetted) error {

	set.contextWasSet = true

	return nil
}
