package main

import (
	"encoding/csv"
	"flag"
	"github.com/moovweb/gokogiri"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

var sem = make(chan int, 10)

type Mod struct {
	name, minecraftVersion string
	currentlyInstalled     string
	curseForgeURL          string
}

type Release struct {
	mod                                               *Mod
	maturity, filename, downloadUrl, minecraftVersion string
}

func releasesFor(mod *Mod) (releases []*Release, err error) {
	trueUrl, e := url.Parse(mod.curseForgeURL)
	if e != nil {
		err = e
		return
	}
	// fetch and read a web page
	resp, e := http.Get(mod.curseForgeURL)
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
		var release Release
		release.mod = mod
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
		downloadUrls, e := file.Search("./td[contains(@class,\"project-file-name\")]//a[contains(@class,\"overflow-tip\")]/@href")
		if e != nil {
			err = e
			return
		}
		for _, downloadUrl := range downloadUrls {
			localUrl, e := url.Parse(downloadUrl.String())
			if e != nil {
				err = e
				return
			}
			release.downloadUrl = trueUrl.ResolveReference(localUrl).String()
			break
		}
		versions, e := file.Search("./td[contains(@class,\"project-file-game-version\")]//span[contains(@class,\"version-label\")]/text()")
		if e != nil {
			err = e
			return
		}
		for _, version := range versions {
			release.minecraftVersion = version.String()
			break
		}
		releases = append(releases, &release)
	}
	return
}
func reportOn(mod *Mod, out chan *Release) (err error) {
	sem <- 1
	releases, e := releasesFor(mod)
	if e != nil {
		err = e
		return
	}
	for _, r := range releases {
		if r.minecraftVersion != mod.minecraftVersion {
			continue
		}
		if r.filename == mod.currentlyInstalled {
			break
		}

		out <- r
	}
	<-sem
	return
}

func main() {
	flag.Parse()
	var input io.Reader
	if flag.NArg() > 1 {
		log.Fatal("Too many arguments")
	} else if flag.NArg() == 1 {
		var err error
		input, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		input = os.Stdin
	}
	reader := csv.NewReader(input)
	writer := csv.NewWriter(os.Stdout)
	reader.FieldsPerRecord = 4
	updates := make(chan *Release)
	var wg sync.WaitGroup
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			wg.Add(1)
			mod := &Mod{name: record[0], minecraftVersion: record[1], currentlyInstalled: record[2], curseForgeURL: record[3]}
			go func() {
				defer wg.Done()
				err := reportOn(mod, updates)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}
	go func() {
		wg.Wait()
		close(updates)
	}()
	for {
		version, ok := <-updates
		if (!ok) {
			break
		}
		tuple := [...]string{version.mod.name, version.minecraftVersion, version.filename, version.mod.curseForgeURL, version.downloadUrl, version.maturity}
		err := writer.Write(tuple[:])
		if err != nil {
			log.Fatal(err)
		}
		writer.Flush()
	}
}
