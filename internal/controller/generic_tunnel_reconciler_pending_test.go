package controller

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestMapSetupNewTunnelError(t *testing.T) {
	t.Run("maps lifecycle pending error to requeue without error", func(t *testing.T) {
		res, err := mapSetupNewTunnelError(&lifecyclePendingError{tunnelName: "k8s-tunnel"})

		assert.NoError(t, err)
		assert.Equal(t, ctrl.Result{RequeueAfter: tunnelLifecycleCheckInterval}, res)
	})

	t.Run("keeps non pending errors unchanged", func(t *testing.T) {
		expectedErr := errors.New("boom")

		res, err := mapSetupNewTunnelError(expectedErr)

		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, ctrl.Result{}, res)
	})
}
