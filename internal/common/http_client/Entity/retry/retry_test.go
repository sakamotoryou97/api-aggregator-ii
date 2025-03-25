package retry_test

import (
	"testing"

	"github.com/sakamotoryou/api-agg-two/internal/common/http_client/Entity/retry"
	"github.com/stretchr/testify/assert"
)

func TestSimpleRetry(t *testing.T) {
	t.Run("two max retry and zero interval", func(t *testing.T) {
    simple := retry.Simple(2, 0, func(code int) bool { return false })
    r, err := retry.New(simple...)

    assert.NoError(t, err)

    counter := 0
    for r.Next() {
      counter++
      r.ValidateCode(400)
    }

    assert.Equal(t, counter, 2)
	})
}
