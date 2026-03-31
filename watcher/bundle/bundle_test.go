package bundle_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/crec-sdk-ext-dvp/watcher/bundle"
)

func TestBundle_Get(t *testing.T) {
	b := bundle.Get()

	t.Run("Service", func(t *testing.T) {
		assert.Equal(t, "dvp", b.Service)
	})

	t.Run("Contracts", func(t *testing.T) {
		require.Len(t, b.Contracts, 1)
		assert.Equal(t, "CCIPDVPCoordinator", b.Contracts[0].Name)
		assert.NotEmpty(t, b.Contracts[0].ABI)
	})

	t.Run("Events", func(t *testing.T) {
		require.NotEmpty(t, b.Events)
		for _, evt := range b.Events {
			assert.NotEmpty(t, evt.Name)
			assert.NotEmpty(t, evt.TriggerContract)
		}
	})
}

func TestBundle_NoDuplicateEvents(t *testing.T) {
	b := bundle.Get()
	seen := make(map[string]bool)
	for _, evt := range b.Events {
		assert.False(t, seen[evt.Name], "duplicate event %q in events list", evt.Name)
		seen[evt.Name] = true
	}
}

func TestBundle_ParamsSchemas(t *testing.T) {
	b := bundle.Get()
	for _, evt := range b.Events {
		if evt.ParamsSchema != nil {
			var schema map[string]any
			require.NoError(t, json.Unmarshal(evt.ParamsSchema, &schema), "event %q has invalid ParamsSchema JSON", evt.Name)
			assert.Contains(t, schema, "properties", "event %q schema should have properties", evt.Name)
		}
	}
}
