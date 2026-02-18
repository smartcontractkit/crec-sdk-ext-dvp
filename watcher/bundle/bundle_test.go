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

	t.Run("ConfigTemplate", func(t *testing.T) {
		assert.NotEmpty(t, b.ConfigTemplate)
	})

	t.Run("Contracts", func(t *testing.T) {
		require.Len(t, b.Contracts, 1)
		assert.Equal(t, "CCIPDVPCoordinator", b.Contracts[0].Name)
		assert.NotEmpty(t, b.Contracts[0].ABI)

		var abi []json.RawMessage
		require.NoError(t, json.Unmarshal([]byte(b.Contracts[0].ABI), &abi))
		assert.Greater(t, len(abi), 0, "ABI should contain entries")
	})

	t.Run("Events", func(t *testing.T) {
		expectedEvents := []string{
			"SettlementOpened",
			"SettlementAccepted",
			"SettlementClosing",
			"SettlementSettled",
			"SettlementCanceling",
			"SettlementCanceled",
		}

		require.Len(t, b.Events, len(expectedEvents))

		for i, name := range expectedEvents {
			t.Run(name, func(t *testing.T) {
				evt := b.Events[i]
				assert.Equal(t, name, evt.Name)
				assert.Equal(t, "CCIPDVPCoordinator", evt.TriggerContract)
				assert.NotEmpty(t, evt.Description)

				assert.NotNil(t, evt.ParamsSchema, "ParamsSchema should not be nil for %s", name)
				var schema map[string]any
				require.NoError(t, json.Unmarshal(evt.ParamsSchema, &schema))
				assert.Contains(t, schema, "properties")

				assert.NotNil(t, evt.DataSchema, "DataSchema should not be nil for %s", name)
				var dataSchema map[string]any
				require.NoError(t, json.Unmarshal(evt.DataSchema, &dataSchema))
				assert.Contains(t, dataSchema, "properties")
			})
		}
	})

	t.Run("FindTriggerContract", func(t *testing.T) {
		c := b.FindTriggerContract("SettlementAccepted")
		require.NotNil(t, c)
		assert.Equal(t, "CCIPDVPCoordinator", c.Name)

		unknown := b.FindTriggerContract("Unknown")
		assert.Nil(t, unknown)
	})

	t.Run("HasEvent", func(t *testing.T) {
		assert.True(t, b.HasEvent("SettlementOpened"))
		assert.True(t, b.HasEvent("SettlementCanceled"))
		assert.False(t, b.HasEvent("Nonexistent"))
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
