package hk_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/remind101/empire/empire"
	"github.com/remind101/empire/empiretest"
)

func TestCreate(t *testing.T) {
	run(t, []Command{
		{
			"apps",
			"",
		},
		{
			"create acme-inc",
			"Created acme-inc.",
		},
	})
}

func TestApps(t *testing.T) {
	run(t, []Command{
		{
			"create acme-inc",
			"Created acme-inc.",
		},
		{
			"apps",
			"acme-inc      Dec 31 17:01",
		},
	})
}

func TestConfig(t *testing.T) {
	run(t, []Command{
		{
			"create acme-inc",
			"Created acme-inc.",
		},
		{
			"set RAILS_ENV=production -a acme-inc",
			"Set env vars and restarted acme-inc.",
		},
		{
			"env -a acme-inc",
			"RAILS_ENV=production",
		},
		{
			"set DATABASE_URL=postgres://localhost AUTH=foo -a acme-inc",
			"Set env vars and restarted acme-inc.",
		},
		{
			"unset RAILS_ENV -a acme-inc",
			"Unset env vars and restarted acme-inc.",
		},
		{
			"env -a acme-inc",
			"AUTH=foo\nDATABASE_URL=postgres://localhost",
		},
	})
}

func TestUpdateConfigNewReleaseSameFormation(t *testing.T) {
	now(time.Now().AddDate(0, 0, -5))
	defer resetNow()

	run(t, []Command{
		{
			"deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
			"Deployed ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
		{
			"dynos -a acme-inc",
			"acme-inc.1.web.1    unknown   5d  \"./bin/web\"",
		},
		{
			"scale web=2 -a acme-inc",
			"Scaled acme-inc to web=2:1X.",
		},
		{
			"dynos -a acme-inc",
			`acme-inc.1.web.1    unknown   5d  "./bin/web"
acme-inc.1.web.2    unknown   5d  "./bin/web"`,
		},
		{
			"set DATABASE_URL=postgres://localhost AUTH=foo -a acme-inc",
			"Set env vars and restarted acme-inc.",
		},
		{
			"dynos -a acme-inc",
			`acme-inc.1.web.1    unknown   5d  "./bin/web"
acme-inc.1.web.2    unknown   5d  "./bin/web"
acme-inc.2.web.1    unknown   5d  "./bin/web"
acme-inc.2.web.2    unknown   5d  "./bin/web"`,
		},
	})
}

func TestDeploy(t *testing.T) {
	run(t, []Command{
		{
			"deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
			"Deployed ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
		{
			"releases -a acme-inc",
			"v1    Dec 31 17:01  Deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
		{
			"deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
			"Deployed ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
		{
			"releases -a acme-inc",
			"v1    Dec 31 17:01  Deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02\nv2    Dec 31 17:01  Deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
	})
}

func TestScale(t *testing.T) {
	now(time.Now().AddDate(0, 0, -5))
	defer resetNow()

	run(t, []Command{
		{
			"deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
			"Deployed ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
		{
			"scale web=2 -a acme-inc",
			"Scaled acme-inc to web=2:1X.",
		},
		{
			"dynos -a acme-inc",
			`acme-inc.1.web.1    unknown   5d  "./bin/web"
acme-inc.1.web.2    unknown   5d  "./bin/web"`,
		},

		{
			"scale web=1 -a acme-inc",
			"Scaled acme-inc to web=1:1X.",
		},
		{
			"dynos -a acme-inc",
			"acme-inc.1.web.1    unknown   5d  \"./bin/web\"",
		},
	})
}

func TestRollback(t *testing.T) {
	run(t, []Command{
		{
			"deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
			"Deployed ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
		{
			"deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
			"Deployed ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02",
		},
		{
			"rollback v1 -a acme-inc",
			"Rolled back acme-inc to v1 as v3.",
		},
		{
			"releases -a acme-inc",
			`v1    Dec 31 17:01  Deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02
v2    Dec 31 17:01  Deploy ejholmes/acme-inc:ec238137726b58285f8951802aed0184f915323668487b4919aff2671c0f9a02
v3    Dec 31 17:01  Rollback to v1`,
		},
	})
}

// Run the tests with empiretest.Run, which will lock access to the database
// since it can't be shared by parallel tests.
func TestMain(m *testing.M) {
	empiretest.Run(m)
}

var fakeNow = time.Date(2015, time.January, 1, 1, 1, 1, 1, time.UTC)

// Stubs out time.Now in empire.
func init() {
	now(fakeNow)
}

// now stubs out empire.Now.
func now(t time.Time) {
	empire.Now = func() time.Time {
		return t
	}
}

func resetNow() {
	now(fakeNow)
}

// hk runs an hk command against a server.
func hk(t testing.TB, url, command string) string {
	args := strings.Split(command, " ")

	cmd := exec.Command("hk", args...)
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		"HKPATH=../../../hk-plugins",
		fmt.Sprintf("HEROKU_API_URL=%s", url),
	}

	b, err := cmd.CombinedOutput()
	t.Log(fmt.Sprintf("\n$ %s\n%s", command, string(b)))
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

type Command struct {
	// Command represents an hk command to run.
	Command string

	// Output is the output we expect to see.
	Output string
}

func run(t testing.TB, commands []Command) {
	s := empiretest.NewServer(t)
	defer s.Close()

	for _, cmd := range commands {
		got := hk(t, s.URL, cmd.Command)

		want := cmd.Output
		if want != "" {
			want = want + "\n"
		}

		if got != want {
			t.Fatalf("%q != %q", got, want)
		}
	}
}