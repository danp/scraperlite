package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/yosssi/gohtml"
)

func main() {
	fs := flag.NewFlagSet("scraperlite", flag.ExitOnError)
	var dbPath string
	fs.StringVar(&dbPath, "db", "data.db", "database file path")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "usage: scraperlite <url> <id1> <css selector1> [<id2> <css selector2> ...]")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Scrape text and HTML based on CSS selectors and save contents to a SQLite database.")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "ids must be in the form:")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "x.txt:  save text of the node specified by the corresponding selector to x")
		fmt.Fprintln(fs.Output(), "x.html: save formatted outer html of the node specified by the corresponding selector to x")
		fmt.Fprintln(fs.Output())
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[1:])

	u := fs.Arg(0)
	if n := fs.NArg() - 1; u == "" || n < 2 || n%2 != 0 {
		fs.Usage()
		os.Exit(1)
	}

	rest := fs.Args()[1:]

	sels := make(map[idType]string)
	var ids []idType
	for len(rest) > 0 {
		rawID, sel := rest[0], rest[1]
		id, typ, ok := strings.Cut(rawID, ".")
		if !ok {
			fs.Usage()
			os.Exit(1)
		}
		switch typ {
		case "txt", "html":
		default:
			fs.Usage()
			os.Exit(1)
		}
		rest = rest[2:]

		idT := idType{id, typ}
		sels[idT] = sel
		ids = append(ids, idT)
	}
	sort.Slice(ids, func(i, j int) bool { return fmt.Sprint(ids[i]) < fmt.Sprint(ids[j]) })

	if err := run(u, dbPath, ids, sels); err != nil {
		log.Fatal(err)
	}
}

type idType struct{ id, typ string }

func run(u, dbPath string, ids []idType, sels map[idType]string) error {
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec("create table if not exists contents (id integer primary key, sha224 text unique, content text)"); err != nil {
		return err
	}
	if _, err := db.Exec("create table if not exists observations (id integer primary key, t datetime, content_id references contents (id))"); err != nil {
		return err
	}

	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		log.Fatal(resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	sum := sha256.New224()

	res := make(map[string]map[string]string)
	for _, id := range ids {
		sel := sels[id]
		docSel := doc.Find(sel)

		var s string
		switch id.typ {
		case "html":
			h, _ := docSel.Parent().Html()
			s = gohtml.Format(h)
		case "txt":
			s = docSel.Text()
		default:
			log.Fatalf("%v missing suffix", id)
		}
		if res[id.id] == nil {
			res[id.id] = make(map[string]string)
		}
		s = strings.TrimSpace(s)
		res[id.id][id.typ] = s

		fmt.Fprintln(sum, id.id, id.typ, s)
	}

	sum64 := base64.RawURLEncoding.EncodeToString(sum.Sum(nil))

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(res); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := db.Exec("insert into contents (sha224, content) values (?, ?) on conflict (sha224) do nothing", sum64, b.Bytes()); err != nil {
		return err
	}
	if _, err := db.Exec("insert into observations (t, content_id) values (?, (select id from contents where sha224=?))", time.Now(), sum64); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
