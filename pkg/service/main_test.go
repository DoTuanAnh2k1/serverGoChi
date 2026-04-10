package service_test

import (
	"os"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

func TestMain(m *testing.M) {
	testutil.InitTestLogger()
	os.Exit(m.Run())
}
