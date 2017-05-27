package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
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

func main() {
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
			// we only want the posts for now
			if strings.Contains(v.Term, "kind#post") {
				fmt.Printf("%s (%s)\n", blogEntry.Title, blogEntry.PublishDate)
				fmt.Printf("\t%s\n\n", blogEntry.Content)
			}
		}
	}
}
