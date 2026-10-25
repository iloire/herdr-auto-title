package resolver

import (
	"strings"
	"testing"

	"github.com/kryptamine/herdr-auto-title/internal/state"
)

func TestIconForWhatAPaneRuns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pane *state.PaneState
		want string
	}{
		{"no pane at all", nil, iconShell},
		{"a bare shell", &state.PaneState{Processes: running("zsh")}, iconShell},
		{
			"an agent, whatever runs beside it",
			&state.PaneState{Agent: "claude", Processes: running("nvim")},
			iconAgent,
		},
		{
			"a known program under its shell",
			&state.PaneState{Processes: running("fish", "nvim")},
			iconVim,
		},
		{"a name in another case", &state.PaneState{Processes: running("LazyGit")}, iconGit},
		{
			"a database client with a mark of its own",
			&state.PaneState{Processes: running("psql")},
			iconPostgres,
		},
		{"a program nobody listed", &state.PaneState{Processes: running("ffmpeg")}, iconProgram},
		// paneKind declines both of these, yet each is still a program.
		{"a runtime", &state.PaneState{Processes: running("node")}, iconNode},
		{"a remote session", &state.PaneState{Processes: running("zsh", "ssh")}, iconRemote},
		// Descendants come before the foreground process, which is listed last.
		{
			"an editor with a language server under it",
			&state.PaneState{Processes: running("node", "nvim")},
			iconVim,
		},
		{"git pushing over ssh", &state.PaneState{Processes: running("ssh", "git")}, iconGit},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := iconFor(tc.pane); got != tc.want {
				t.Errorf("icon = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIconLeadsTheTitle(t *testing.T) {
	t.Parallel()

	r := NewIconedTitles(New(Options{}, NewCWD("")), 0)

	if got := r.Resolve(tabWithCWD(dashboard)); got.Name != iconShell+" dashboard" {
		t.Errorf("name = %q, want %q", got.Name, iconShell+" dashboard")
	}
}

func TestIconSitsInsideThePosition(t *testing.T) {
	// The position is the key that switches to a tab, so it keeps the lead.
	t.Parallel()

	inner := NewIconedTitles(New(Options{}, NewCWD("")), 0)
	tab := atPosition(2, dashboard)

	want := "2 · " + iconShell + " dashboard"
	if got := NewNumbered(inner, 0).Resolve(tab); got.Name != want {
		t.Errorf("name = %q, want %q", got.Name, want)
	}
}

func TestIconCountsAgainstTheLength(t *testing.T) {
	// An icon must not make a title wider than the bound the user set.
	t.Parallel()

	const maxLength = 6

	r := NewIconedTitles(New(Options{MaxLength: maxLength}, NewCWD("")), maxLength)

	want := iconShell + " das…"
	if got := r.Resolve(tabWithCWD(dashboard)); got.Name != want {
		t.Errorf("name = %q, want %q", got.Name, want)
	}
}

func TestEachPaneTakesItsOwnIcon(t *testing.T) {
	t.Parallel()

	tab := tabOf([]*state.PaneState{
		{ID: "wE:p1", Dir: dashboard, Focused: true, Processes: running("nvim")},
		{ID: "wE:p2", Dir: dashboard, Agent: "claude"},
	})

	decisions := NewIconedPanes(Default(Options{}), 0).ResolvePanes(tab)

	for i, want := range []string{iconVim, iconAgent} {
		if got := decisions[i].Name; !strings.HasPrefix(got, want+" ") {
			t.Errorf("pane %d = %q, want it to lead with %q", i, got, want)
		}
	}
}
