package internal_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/danp/scraperlite/internal"
	"github.com/google/go-cmp/cmp"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	d := t.TempDir()

	start := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start

	s := "1"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
<ul>
<li>one</li>
<li>two</li>
</ul>
`)

		fmt.Fprintln(w, "<span id=\"s\">", s, "</span>")
	}))
	defer ts.Close()

	dbPath := filepath.Join(d, "test.db")

	args := []string{"scraperlite", "-db", dbPath, ts.URL, "ul.txt", "ul", "ul.html", "ul", "li.html", "li:nth-child(2)", "s.txt", "#s"}

	var b bytes.Buffer

	run := func() {
		t.Helper()
		if err := internal.Run(args, &b, func() time.Time { return now }); err != nil {
			t.Fatal(err)
		}
	}

	run()

	now = now.Add(time.Second)

	run()

	now = now.Add(time.Second)

	s = "2"

	run()

	now = now.Add(time.Second)

	s = "1"

	run()

	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_pragma=foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("select t, contents.id, content from observations join contents on content_id = contents.id")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type ObservationContent struct {
		T         time.Time
		ContentID int
		Content   string
	}

	var obs []ObservationContent
	for rows.Next() {
		var o ObservationContent
		if err := rows.Scan(&o.T, &o.ContentID, &o.Content); err != nil {
			t.Fatal(err)
		}
		obs = append(obs, o)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []ObservationContent{
		{
			T:         start,
			ContentID: 1,
			Content: `{"li":{"html":"<li>\n  two\n</li>"},"s":{"txt":"1"},"ul":{"html":"<ul>\n  <li>\n    one\n  </li>\n  <li>\n    two\n  </li>\n</ul>","txt":"one\ntwo"}}
`,
		},
		{
			T:         start.Add(time.Second),
			ContentID: 1,
			Content: `{"li":{"html":"<li>\n  two\n</li>"},"s":{"txt":"1"},"ul":{"html":"<ul>\n  <li>\n    one\n  </li>\n  <li>\n    two\n  </li>\n</ul>","txt":"one\ntwo"}}
`,
		},
		{
			T:         start.Add(2 * time.Second),
			ContentID: 2,
			Content: `{"li":{"html":"<li>\n  two\n</li>"},"s":{"txt":"2"},"ul":{"html":"<ul>\n  <li>\n    one\n  </li>\n  <li>\n    two\n  </li>\n</ul>","txt":"one\ntwo"}}
`,
		},
		{
			T:         start.Add(3 * time.Second),
			ContentID: 1,
			Content: `{"li":{"html":"<li>\n  two\n</li>"},"s":{"txt":"1"},"ul":{"html":"<ul>\n  <li>\n    one\n  </li>\n  <li>\n    two\n  </li>\n</ul>","txt":"one\ntwo"}}
`,
		},
	}
	if d := cmp.Diff(want, obs); d != "" {
		t.Fatal(d)
	}

	if got := b.String(); got != "" {
		t.Fatal(got)
	}
}
