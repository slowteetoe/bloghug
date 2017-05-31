package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	html "html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type feed struct {
	ID         string  `xml:"id"`
	Title      string  `xml:"title"`
	AuthorName string  `xml:"author>name"`
	Entries    []entry `xml:"entry"`
}

// Since we can't use time.Time directly, create custom
// type with String and UnmarshalXML methods
type xmlDate struct {
	time.Time
}

func (x *xmlDate) String() string {
	return x.Format("2006-01-02T15:04:05Z")
}

func (x *xmlDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse("2006-01-02T15:04:05.000-07:00", v)
	if err != nil {
		return err
	}
	*x = xmlDate{parse}
	return nil
}

type entry struct {
	PublishDate xmlDate    `xml:"published"`
	Categories  []category `xml:"category"`
	Title       string     `xml:"title"`
	Content     string     `xml:"content"`
}

type category struct {
	Term  string `xml:"term,attr"`
	Title string `xml:"title"`
}

var (
	outputDir = flag.String("outputDir", "./content/", "output directory for files, e.g. /fully/qualified/content/")
	ltRe      = regexp.MustCompile("&lt;")
	gtRe      = regexp.MustCompile("&gt;")
)

func main() {
	flag.Parse()

	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		fmt.Printf("cannot use specified output directory: %v\n", err)
		return
	}
	funcMap := template.FuncMap{"renderSafe": renderSafe}
	t, err := template.New("output.tmpl").Funcs(funcMap).ParseFiles("templates/output.tmpl")
	if err != nil {
		fmt.Printf("unable to parse template: %v\n", err)
		return
	}

	fmt.Println("Converting blogger to hugo format")

	xmlFile, err := os.Open("data/blog.xml")
	if err != nil {
		fmt.Printf("unable to open file: %v\n", err)
		return
	}
	defer xmlFile.Close()

	var f feed
	if err = xml.NewDecoder(xmlFile).Decode(&f); err != nil {
		fmt.Printf("unable to decode xml: %v\n", err)
		return
	}

	fmt.Printf("Converting '%s' - by %s\n", f.Title, f.AuthorName)

	for _, blogEntry := range f.Entries {
		// everything is lumped together under the <entry> tag
		for _, v := range blogEntry.Categories {
			// we only want the posts for now, maybe comments later...
			if strings.Contains(v.Term, "kind#post") {
				filename := toFilename(blogEntry.Title)
				fmt.Printf("Writing to: %s\n", filename)
				f, err := os.OpenFile(filepath.Join(*outputDir, filename), os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Printf("unable to open file to write blog entry: %v\n", err)
					return
				}
				defer f.Close()

				blogEntry.Content = gtRe.ReplaceAllString(ltRe.ReplaceAllString(blogEntry.Content, "<"), ">")

				// remove the http://schemas.google.com/blogger/2008/kind#post category
				cleanCats := []category{}
				for i := range blogEntry.Categories {
					if !strings.Contains(blogEntry.Categories[i].Term, "kind#post") {
						cleanCats = append(cleanCats, blogEntry.Categories[i])
					}
				}
				blogEntry.Categories = cleanCats

				if err = t.Execute(f, blogEntry); err != nil {
					fmt.Printf("unable to write template: %v\n", err)
					return
				}
				break
			}
		}
	}
}

var re = regexp.MustCompile("[^a-zA-Z0-9-_]+")

func toFilename(s string) string {
	return fmt.Sprintf("%s.md", strings.ToLower(re.ReplaceAllString(s, "-")))
}

func renderSafe(s string) html.HTML {
	return html.HTML(s)
}
