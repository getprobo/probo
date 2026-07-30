package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyOptions(t *testing.T) {
	t.Run("no options uses default recover func", func(t *testing.T) {
		opts := ApplyOptions(nil)
		assert.NotNil(t, opts.RecoverFunc)
	})

	t.Run("nil recover func falls back to default", func(t *testing.T) {
		opts := ApplyOptions([]Option{WithRecoverFunc(nil)})
		assert.NotNil(t, opts.RecoverFunc)
	})

	t.Run("with custom recover func", func(t *testing.T) {
		fn := func(_ context.Context, _ any) error {
			return errors.New("sanitized")
		}
		opts := ApplyOptions([]Option{WithRecoverFunc(fn)})
		assert.NotNil(t, opts.RecoverFunc)

		err := opts.RecoverFunc(context.Background(), "boom")
		assert.Equal(t, "sanitized", err.Error())
	})

	t.Run("recover func receives raw panic value", func(t *testing.T) {
		var captured any
		fn := func(_ context.Context, err any) error {
			captured = err
			return nil
		}
		opts := ApplyOptions([]Option{WithRecoverFunc(fn)})

		opts.RecoverFunc(context.Background(), 42)
		assert.Equal(t, 42, captured)

		opts.RecoverFunc(context.Background(), "string panic")
		assert.Equal(t, "string panic", captured)

		original := errors.New("error panic")
		opts.RecoverFunc(context.Background(), original)
		assert.Equal(t, original, captured)
	})
}
