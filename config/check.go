package config

import (
	"errors"
	"fmt"

	"github.com/Sn0wo2/CatSync/config/reader"
	"go.uber.org/zap"
)

type validationErrors struct {
	err []error
}

func (e *validationErrors) add(err error) {
	if err != nil {
		e.err = append(e.err, err)
	}
}

func validateRequiredString(where string, value *reader.String) error {
	if value == nil {
		return fmt.Errorf("%s is required", where)
	}

	if err := value.ValidateNoIO(); err != nil {
		return fmt.Errorf("invalid %s: %w", where, err)
	}

	return nil
}

func validateOptionalString(where string, value *reader.String) error {
	if value == nil {
		return nil
	}

	if content, literal := value.LiteralTrim(); literal && content == "" {
		return nil
	}

	if err := value.ValidateNoIO(); err != nil {
		return fmt.Errorf("invalid %s: %w", where, err)
	}

	return nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("nil config")
	}

	ec := &validationErrors{}
	addErr := ec.add

	addErr(validateRequiredString("log.fileFormat", c.Log.FileFormat))
	addErr(validateRequiredString("server.address", c.Server.Address))

	if len(c.Actions) == 0 {
		addErr(errors.New("actions is required"))
	}

	for i := range c.Actions {
		if c.Actions[i].Type == "" {
			addErr(fmt.Errorf("actions[%d].type is required", i))
		}
	}

	c.checkACME(addErr)
	c.checkGlobalModifiers(addErr)
	c.checkActions(addErr)

	return errors.Join(ec.err...)
}

func (c *Config) LogWarnings(logger *zap.Logger) {
	if c == nil || logger == nil {
		return
	}

	if len(c.Actions) == 0 {
		logger.Warn("Config >> no actions configured; router will fall through to fiber (ctx.Next())")

		return
	}

	for i := range c.Actions {
		action := &c.Actions[i]

		route, _ := action.Route.LiteralTrim()
		if route == "" {
			logger.Info("Config >> action route is empty; action is jump-only",
				zap.Int("index", i),
				zap.String("type", string(action.Type)),
			)
		}
	}

	lastIndex := len(c.Actions) - 1
	last := c.Actions[lastIndex]
	logger.Info("Config >> notfound handler is the last action",
		zap.Int("index", lastIndex),
		zap.String("type", string(last.Type)),
	)

	if last.Type == ActionFile {
		logger.Warn("Config >> notfound handler is file action; may leak file contents",
			zap.Int("index", lastIndex),
			zap.String("type", string(last.Type)),
		)
	}
}
