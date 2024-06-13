package run

import (
	"context"
	"testing"

	savvy_client "github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/idgen"
	"github.com/stretchr/testify/assert"
)

func TestRunServer(t *testing.T) {
	rb := &savvy_client.Runbook{
		Title: "test",
		Steps: []savvy_client.Step{
			{
				Command: "idx_0",
			},
			{
				Command: "idx_1",
			},
			{
				Command: "idx_2",
			},
		},
	}

	socketPath := "/tmp/savvy-run-test-" + idgen.New("tst") + ".sock"

	srv, err := NewServerWithSocketPath(socketPath, rb)
	// test server setup
	assert.Nil(t, err)
	assert.NotNil(t, srv)
	assert.Equal(t, socketPath, srv.SocketPath())
	assert.Len(t, srv.Commands(), 3)

	assert.Equal(t, "idx_0", srv.Commands()[0].Command)
	assert.Equal(t, "idx_1", srv.Commands()[1].Command)
	assert.Equal(t, "idx_2", srv.Commands()[2].Command)

	ctx := context.Background()
	cl, err := NewClient(ctx, srv.SocketPath())
	assert.Nil(t, err)
	assert.NotNil(t, cl)

	t.Cleanup(func() { assert.NoError(t, srv.Close()) })

	go srv.ListenAndServe()

	t.Run("TestCurrentCommand", func(t *testing.T) {
		// test current command
		st, err := cl.CurrentState()
		assert.NoError(t, err)
		assert.NotNil(t, st)
		assert.Equal(t, "idx_0", st.Command)

		t.Run("TestCurrentCommandIdempotent", func(t *testing.T) {
			// test current command
			st, err := cl.CurrentState()
			assert.NoError(t, err)
			assert.NotNil(t, st)
			assert.Equal(t, "idx_0", st.Command)
		})
	})
	t.Run("TestNextCommand", func(t *testing.T) {
		st, err := cl.CurrentState()
		assert.NoError(t, err)
		assert.NotNil(t, st)
		assert.Zero(t, st.Index)

		assert.NoError(t, cl.NextCommand())

		st, err = cl.CurrentState()
		assert.NoError(t, err)
		assert.NotNil(t, st)
		assert.Equal(t, 1, st.Index)
		assert.Equal(t, "idx_1", st.Command)
	})
}