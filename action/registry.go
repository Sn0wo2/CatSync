package action

import "github.com/Sn0wo2/CatSync/config"

// Registry maps action types to their handlers.
// Use NewRegistry() to create a registry with all built-in handlers.
type Registry map[config.ActionType]Handler

func NewRegistry() Registry {
	return Registry{
		config.ActionString: NewString(),
		config.ActionFile:   NewFile(),
		config.ActionServer: NewServer(),
		config.ActionReload: NewReload(),
	}
}
