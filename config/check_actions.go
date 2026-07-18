package config

import (
	"fmt"

	"github.com/Sn0wo2/CatSync/internal/util"
)

func (c *Config) checkActions(add func(error)) {
	actionCount := len(c.Actions)

	labelIndex := make(map[string]int, actionCount)

	for i, act := range c.Actions {
		if act.Label != "" {
			if _, dup := labelIndex[act.Label]; dup {
				add(fmt.Errorf("duplicate action label at actions[%d]: %q", i, act.Label))
			} else {
				labelIndex[act.Label] = i
			}
		}
	}

	for i, act := range c.Actions {
		route, ok := act.Route.LiteralTrim()
		if !ok {
			add(fmt.Errorf("actions[%d].route must be a literal string (type=string)", i))

			route = ""
		}

		if route != "" {
			if _, err := util.GetCompiledRegexp(route); err != nil {
				add(fmt.Errorf("invalid action route regexp at actions[%d].route (%q): %w", i, route, err))
			}
		}

		c.validateModifier(fmt.Sprintf("actions[%d]", i), &act.GlobalModifier, actionCount, labelIndex, add)

		payload := act.GetPayload()
		if payload == nil {
			add(fmt.Errorf("actions[%d] type=%s but payload is nil", i, act.Type))

			continue
		}

		switch act.Type {
		case ActionFile:
			add(validateRequiredString(fmt.Sprintf("actions[%d].file.path", i), act.ActionFile.Path))
		case ActionString:
			add(validateOptionalString(fmt.Sprintf("actions[%d].string.content", i), act.ActionString.Content))
		case ActionServer:
			add(validateRequiredString(fmt.Sprintf("actions[%d].server.directory", i), act.ActionServer.Directory))

			for j, indexFile := range act.ActionServer.IndexFiles {
				add(validateRequiredString(fmt.Sprintf("actions[%d].server.indexFiles[%d]", i, j), indexFile))
			}
		case ActionReload:
		}

		c.validateModifier(fmt.Sprintf("actions[%d].%s", i, act.Type), payload.GetGlobalModifier(), actionCount, labelIndex, add)
	}
}
