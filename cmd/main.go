package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

type building struct {
	name      string
	architect string
	city      string
	state     string
	country   string
	latitude  string
	longitude string
	date      string
	style     string
	_type     string
	alias     string
	notes     string
}

func main() {
	buildings := []string{
		"http://www.worldarchitecturemap.org/buildings/0-14",                  //with notes
		"http://www.worldarchitecturemap.org/buildings/100-social-apartments", //Without notes
		"http://www.worldarchitecturemap.org/buildings/bkkjnlknjk",            //None existing building
	}
	for _, buildingUrl := range buildings {
		fmt.Println("\nParsing building: " + buildingUrl)
		b, err := parseBuilding(buildingUrl)
		if err == nil {
			fmt.Printf("%v", b)
		} else {
			fmt.Printf(err.Error())
		}
	}
}

func parseBuilding(url string) (building, error) {

	var b building

	resp, err := http.Get(url)

	if err != nil {
		return b, err
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("unexpected response from the web. HTTP code: %d", resp.StatusCode)
		return b, errors.New(msg)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	buildingInfo := doc.Find(".building_info")

	b.name = buildingInfo.Find("h1").Text()

	tableHtml := buildingInfo.Find("#building_info_tbl").Find("tbody")

	//Architect row
	architectRow := tableHtml.Find("tr")
	b.architect = architectRow.Find("a").First().Text()

	//Location row
	locationRow := architectRow.Siblings().First()
	locationRow.Find("a").Each(func(i int, selection *goquery.Selection) {
		switch i {
		case 0:
			b.city = selection.Text()
		case 1:
			b.state = selection.Text()
		case 2:
			b.country = selection.Text()
		}
	})

	//GPS row
	gpsRow := tableHtml.Find("tr").Eq(2)
	gpsHTML, err := gpsRow.Html()

	if err != nil {
		fmt.Printf("Error parsing GPS HTML: %s", err.Error())
	} else {
		b.latitude = getStringInBetween(gpsHTML, "(", ")")
		pos := strings.Index(gpsHTML, ",")
		b.longitude = getStringInBetween(gpsHTML[pos:], "(", ")")
	}

	//Date row
	dateRow := tableHtml.Find("tr").Eq(3)
	b.date = dateRow.Children().Next().Text()

	//Style row
	styleRow := tableHtml.Find("tr").Eq(4)
	b.style = styleRow.Find("a").Text()

	//Type row
	typeRow := tableHtml.Find("tr").Eq(5)
	b._type = typeRow.Find("a").Text()

	//Alias row
	aliasRow := tableHtml.Find("tr").Eq(6)
	b.alias = aliasRow.Children().Next().Text()

	//Notes row
	notesRow := tableHtml.Find("tr").Eq(7)
	b.notes = notesRow.Children().Next().Text()

	//Check building existence
	if b.name == "" && b.architect == "" {
		return b, errors.New("the building doesn't exist")
	}

	return b, nil
}

// getStringInBetween returns empty string if no start or end string found
func getStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}
