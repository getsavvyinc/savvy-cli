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

	t.Run("TestCurrentCommand", func(t *testing.T) {
		srv, cl, cleanup := newTestServerWithClient(t, rb)
		t.Cleanup(func() { cleanup() })

		assert.Len(t, srv.Commands(), 3)
		assert.Equal(t, "idx_0", srv.Commands()[0].Command)
		assert.Equal(t, "idx_1", srv.Commands()[1].Command)
		assert.Equal(t, "idx_2", srv.Commands()[2].Command)
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
		_, cl, cleanup := newTestServerWithClient(t, rb)
		t.Cleanup(func() { cleanup() })

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
	t.Run("TestParam", func(t *testing.T) {
		_, cl, cleanup := newTestServerWithClient(t, rb)
		t.Cleanup(func() { cleanup() })

		st, err := cl.CurrentState()
		assert.NoError(t, err)
		assert.NotNil(t, st)
		assert.Zero(t, st.Index)
		assert.Len(t, st.Params, 0)
		t.Run("TestSetParam", func(t *testing.T) {
			assert.NoError(t, cl.SetParams(map[string]string{"<param>": "value"}))
			st, err := cl.CurrentState()
			assert.NoError(t, err)
			assert.NotNil(t, st)
			assert.Zero(t, st.Index)
			assert.Len(t, st.Params, 1)
			assert.Equal(t, "value", st.Params["<param>"])
			t.Run("TestNoOverwriteParam", func(t *testing.T) {
				assert.NoError(t, cl.SetParams(map[string]string{"<param>": "anotherValue"}))
				st, err := cl.CurrentState()
				assert.NoError(t, err)
				assert.NotNil(t, st)
				assert.Zero(t, st.Index)
				assert.Len(t, st.Params, 1)
				assert.Equal(t, "value", st.Params["<param>"])
			})
			t.Run("TestParamStateIsMaintained", func(t *testing.T) {
				t.Run("WithNextCommand", func(t *testing.T) {
					assert.NoError(t, cl.NextCommand())
					st, err := cl.CurrentState()
					assert.NoError(t, err)
					assert.NotNil(t, st)
					assert.Equal(t, 1, st.Index)
					assert.Equal(t, "idx_1", st.Command)
					assert.Len(t, st.Params, 1)
					assert.Equal(t, "value", st.Params["<param>"])
				})
				t.Run("WithMoreParamsAdded", func(t *testing.T) {
					assert.NoError(t, cl.SetParams(map[string]string{"<param2>": "value2"}))
					st, err := cl.CurrentState()
					assert.NoError(t, err)
					assert.NotNil(t, st)
					assert.Equal(t, st.Index, 1)
					assert.Len(t, st.Params, 2)
					assert.Equal(t, "value", st.Params["<param>"])
					assert.Equal(t, "value2", st.Params["<param2>"])
				})
			})
		})

	})
}

type cleanupFunc func() error

func newTestServerWithClient(t *testing.T, rb *savvy_client.Runbook) (*RunServer, Client, cleanupFunc) {
	socketPath := "/tmp/savvy-run-test-" + idgen.New("tst") + ".sock"

	srv, err := NewServerWithSocketPath(socketPath, rb)
	assert.Nil(t, err)
	assert.NotNil(t, srv)
	assert.Equal(t, socketPath, srv.SocketPath())

	ctx := context.Background()
	cl, err := NewClient(ctx, srv.SocketPath())
	assert.Nil(t, err)
	assert.NotNil(t, cl)

	go srv.ListenAndServe()
	return srv, cl, srv.Close
}
