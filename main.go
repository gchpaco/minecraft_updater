package main

import (
	"github.com/moovweb/gokogiri"
	"io/ioutil"
	"net/http"
	"fmt"
	"os"
	"text/tabwriter"
)

func main() {
	// fetch and read a web page
	resp, _ := http.Get("http://minecraft.curseforge.com/mc-mods/231549-compacter/files")
	page, _ := ioutil.ReadAll(resp.Body)

	// parse the web page
	doc, _ := gokogiri.ParseHtml(page)

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 1, ' ', 0)

	html := doc.Root().FirstChild()
	files, _ := html.Search("//tr[contains(@class,\"project-file-list-item\")]")

	for _, file := range files {
		releaseTypes, _ := file.Search("./td[contains(@class,\"project-file-release-type\")]/div/@title")
		fmt.Fprint(w, "Release:")
		for _, releaseType := range releaseTypes {
			fmt.Fprint(w, "\t", releaseType)
		}
		filenames, _ := file.Search("./td[contains(@class,\"project-file-name\")]//a[contains(@class,\"overflow-tip\")]/text()")
		for _, filename := range filenames {
			fmt.Fprint(w, "\t", filename)
		}
		fmt.Fprintln(w)
	}
	w.Flush()

	// perform operations on the parsed page -- consult the tests for examples

	// important -- don't forget to free the resources when you're done!
	doc.Free()
}
