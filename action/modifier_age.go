package action

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/gofiber/fiber/v3"
)

type AgeModifier struct {
	recipients []age.Recipient
	armor      bool
}

func NewAgeModifier() *AgeModifier {
	return &AgeModifier{armor: true}
}

func (m *AgeModifier) WithRecipients(recipients []age.Recipient) *AgeModifier {
	m.recipients = recipients

	return m
}

func (m *AgeModifier) WithArmor(armor bool) *AgeModifier {
	m.armor = armor

	return m
}

func (m *AgeModifier) Before(p *ProcessData) (*ProcessData, ExecutionResult) {
	return p, ExecutionResult{}
}

func (m *AgeModifier) After(p *ProcessData) (*ProcessData, ExecutionResult) {
	body := p.FCtx.Response().Body()
	if len(body) == 0 {
		return p, ExecutionResult{}
	}

	if len(m.recipients) == 0 {
		return nil, ExecutionResult{Err: errors.New("age modifier requires at least one recipient")}
	}

	var buf bytes.Buffer

	var (
		dst    io.Writer = &buf
		closer io.Closer
	)

	if m.armor {
		aw := armor.NewWriter(&buf)
		dst, closer = aw, aw
	} else {
		p.FCtx.Set(fiber.HeaderContentType, "application/octet-stream")
	}

	w, err := age.Encrypt(dst, m.recipients...)
	if err != nil {
		return nil, ExecutionResult{Err: fmt.Errorf("age encrypt error: %w", err)}
	}

	if _, err := w.Write(body); err != nil {
		return nil, ExecutionResult{Err: fmt.Errorf("age write error: %w", err)}
	}

	if err := w.Close(); err != nil {
		return nil, ExecutionResult{Err: fmt.Errorf("age close error: %w", err)}
	}

	if closer != nil {
		if err := closer.Close(); err != nil {
			return nil, ExecutionResult{Err: fmt.Errorf("age armor close error: %w", err)}
		}

		p.FCtx.Set(fiber.HeaderContentType, "text/plain; charset=utf-8")
	}

	p.FCtx.Response().SetBody(buf.Bytes())

	return p, ExecutionResult{}
}
