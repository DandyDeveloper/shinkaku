package importer

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/nihongo-sensei/backend/internal/db"
)

// JMdictImporter parses a JMdict XML (or .xml.gz) file and imports only
// "common" entries — those with a ke_pri or re_pri value of ichi1, ichi2,
// news1, news2, or nf01–nf24.
//
// Download from https://www.edrdg.org/jmdict/edict_doc.html
type JMdictImporter struct{}

func (j *JMdictImporter) Source() string { return "jmdict" }

type jmEntry struct {
	Seq   string    `xml:"ent_seq"`
	KEle  []jmKEle  `xml:"k_ele"`
	REle  []jmREle  `xml:"r_ele"`
	Sense []jmSense `xml:"sense"`
}

type jmKEle struct {
	Keb   string   `xml:"keb"`
	KePri []string `xml:"ke_pri"`
}

type jmREle struct {
	Reb   string   `xml:"reb"`
	RePri []string `xml:"re_pri"`
}

type jmSense struct {
	Gloss []jmGloss `xml:"gloss"`
}

type jmGloss struct {
	Text string `xml:",chardata"`
	Lang string `xml:"lang,attr"`
}

func jmIsCommon(priorities []string) bool {
	for _, p := range priorities {
		switch p {
		case "ichi1", "ichi2", "news1", "news2":
			return true
		}
		// nf01–nf24 marks the 2400 most common words in newspapers.
		if len(p) == 4 && p[:2] == "nf" {
			return true
		}
	}
	return false
}

func (j *JMdictImporter) Import(ctx context.Context, database *db.DB, opts Options) (Result, error) {
	if opts.File == "" {
		return Result{}, fmt.Errorf("--file is required for jmdict source (download JMdict_e.gz from edrdg.org)")
	}

	f, err := os.Open(opts.File)
	if err != nil {
		return Result{}, fmt.Errorf("open %s: %w", opts.File, err)
	}
	defer f.Close()

	var r io.Reader = f
	if strings.HasSuffix(strings.ToLower(opts.File), ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return Result{}, fmt.Errorf("gzip reader: %w", err)
		}
		defer gz.Close()
		r = gz
	}

	dec := xml.NewDecoder(r)
	dec.Strict = false // JMdict uses custom entities; ignore unknown ones

	var result Result
	for {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, fmt.Errorf("xml token: %w", err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "entry" {
			continue
		}

		var entry jmEntry
		if err := dec.DecodeElement(&entry, &se); err != nil {
			result.Errors++
			continue
		}

		// Keep only common entries.
		common := false
		for _, k := range entry.KEle {
			if jmIsCommon(k.KePri) {
				common = true
				break
			}
		}
		if !common {
			for _, re := range entry.REle {
				if jmIsCommon(re.RePri) {
					common = true
					break
				}
			}
		}
		if !common {
			continue
		}

		if opts.Limit > 0 && result.Imported >= opts.Limit {
			break
		}

		// Prefer kanji form; fall back to kana reading.
		pattern := ""
		if len(entry.KEle) > 0 {
			pattern = entry.KEle[0].Keb
		} else if len(entry.REle) > 0 {
			pattern = entry.REle[0].Reb
		}
		if pattern == "" {
			result.Skipped++
			continue
		}

		// Collect English glosses from all senses.
		var glosses []string
		for _, s := range entry.Sense {
			for _, g := range s.Gloss {
				if g.Lang == "" || g.Lang == "eng" {
					glosses = append(glosses, g.Text)
				}
			}
		}
		meaning := strings.Join(glosses, "; ")
		if len(meaning) > 255 {
			meaning = meaning[:255]
		}

		res, err := database.ExecContext(ctx,
			`INSERT OR IGNORE INTO grammar_points
			    (source, external_id, jlpt_level, pattern, meaning, tags)
			 VALUES (?, ?, '', ?, ?, 'vocabulary')`,
			"jmdict", entry.Seq, pattern, meaning)
		if err != nil {
			result.Errors++
			continue
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			result.Skipped++
		} else {
			result.Imported++
		}
	}
	return result, nil
}
