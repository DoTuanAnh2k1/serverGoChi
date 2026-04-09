package service_test

import (
	"os"
	"testing"

	"go-aa-server/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.InitTestLogger()
	os.Exit(m.Run())
}
