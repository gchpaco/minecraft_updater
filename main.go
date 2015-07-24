package main

import (
	"fmt"
	"github.com/moovweb/gokogiri"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"
)

type Release struct {
	maturity string
	filename string
}

func releasesFor(url string) (releases []Release, err error) {
	// fetch and read a web page
	resp, e := http.Get(url)
	if e != nil {
		err = e
		return
	}
	page, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		err = e
		return
	}

	// parse the web page
	doc, e := gokogiri.ParseHtml(page)
	if e != nil {
		err = e
		return
	}
	defer doc.Free()

	html := doc.Root().FirstChild()
	files, e := html.Search("//tr[contains(@class,\"project-file-list-item\")]")
	if e != nil {
		err = e
		return
	}

	for _, file := range files {
		release := &Release{}
		releaseTypes, e := file.Search("./td[contains(@class,\"project-file-release-type\")]/div/@title")
		if e != nil {
			err = e
			return
		}
		for _, releaseType := range releaseTypes {
			release.maturity = releaseType.String()
			break
		}
		filenames, e := file.Search("./td[contains(@class,\"project-file-name\")]//a[contains(@class,\"overflow-tip\")]/text()")
		if e != nil {
			err = e
			return
		}
		for _, filename := range filenames {
			release.filename = filename.String()
			break
		}
		releases = append(releases, *release)
	}
	return
}
func latestFor(url string) (release Release, err error) {
	releases, e := releasesFor(url)
	if e != nil {
		err = e
		return
	}

	release = releases[0]
	return
}
func reportOn(mod, url string, w io.Writer) (err error) {
	release, e := latestFor(url)
	if e != nil {
		err = e
		return
	}

	fmt.Fprintln(w, mod, "\t", release.maturity, "\t", release.filename)
	return
}

func main() {
	w := tabwriter.NewWriter(os.Stdout, 5, 8, 1, ' ', tabwriter.TabIndent)
	err := reportOn("Compacter", "http://minecraft.curseforge.com/mc-mods/231549-compacter/files", w)
	if err != nil {
		panic(err)
	}
	w.Flush()
}
