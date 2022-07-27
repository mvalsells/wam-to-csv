package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"path"
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

const (
	BUILDING_BASE_URL    = "http://www.worldarchitecturemap.org/buildings/"
	BUILDING_TABLE_DATE  = "Date"
	BUILDING_TABLE_STYLE = "Style"
	BUILDING_TABLE_TYPE  = "Type"
	BUILDING_TABLE_ALIAS = "Alias"
	BUILDING_TABLE_NOTES = "Notes"
)

func main() {

	var allBuildingsUrl []string
	var allBuildings []building

	//Get all the urls
	fmt.Println("Starting to get all buildings urls")
	allBuildingsUrl = append(allBuildingsUrl, parseLetterBuildingList(BUILDING_BASE_URL)...)
	fmt.Printf("Collected all numbers buildings urls, total: %d\n", len(allBuildingsUrl))
	for c := 'a'; c <= 'z'; c++ {
		url := fmt.Sprintf("%s?letter=%s", BUILDING_BASE_URL, string(c))
		bList := parseLetterBuildingList(url)
		fmt.Printf("Collected all letter %s buildings urls, total: %d\n", string(c), len(bList))
		allBuildingsUrl = append(allBuildingsUrl, bList...)
	}
	fmt.Printf("Finished getting all the buildings urls. Total urls collected: %d\n------------------------------------------------------------------------------------\n", len(allBuildingsUrl))

	//Save all the building information
	fmt.Println("\nStarting to download buildings information")
	for i, bUrl := range allBuildingsUrl {
		b, err := parseBuilding(bUrl)
		if err == nil {
			allBuildings = append(allBuildings, b)

		} else {
			fmt.Printf("Error when parsing %s: %s\n", bUrl, err.Error())
		}
		if i%100 == 99 {
			fmt.Printf("Downloaded information from %d/%d buildings\n", i+1, len(allBuildingsUrl))
		}
	}

	//Save data to file
	//dir, err := os.Executable()
	dir := "/home/mvalsells/Desktop"
	var err error = nil
	if err == nil {
		//err = saveBuildingsToCsv(allBuildings, path.Join(dir, "wam-export.csv"))
		err = saveBuildingsToCsv(allBuildings, path.Join(dir, "wam-export.csv"))
		if err != nil {
			fmt.Printf("Unable to write buildings to a file: %s\n", err.Error())
		} else {
			fmt.Printf("Data saved in the %s file.\n", dir)
		}
	} else {
		fmt.Printf("Unable to get current path, data will not be saved to a file\n")
	}
	fmt.Printf("Job finished, exiting...\n")
}

//Given a slice of buildings save them in a CSV format in the provided file path
//Returns error if it wasn't unable to save them in the file
func saveBuildingsToCsv(buildings []building, filePath string) error {
	csvText := []string{`"name", "architect", "city", "state", "country", "latitude", "longitude", "date", "style", "type", "alias", "notes"`}

	for _, b := range buildings {
		s := fmt.Sprintf(`"%s", "%s", "%s", "%s", "%s", "%s", "%s", "%s", "%s", "%s", "%s", "%s"`,
			b.name,
			b.architect,
			b.city,
			b.state,
			b.country,
			b.latitude,
			b.longitude,
			b.date,
			b.style,
			b._type,
			b.alias,
			b.notes)
		csvText = append(csvText, s)
	}

	err := ioutil.WriteFile(filePath, []byte(strings.Join(csvText, "\n")), 0644)
	return err
}

//Given a letter returns all the buildings urls starting with that letter
func parseLetterBuildingList(baseUrl string) []string {

	var letterBuildingList []string
	var tmpList []string
	var err error

	//1st page
	tmpList, err = parsePageBuildingList(baseUrl)
	if err == nil {
		letterBuildingList = append(letterBuildingList, tmpList...)
	} else {
		fmt.Sprintf("Error when parsing %s: %s", baseUrl, err.Error())
	}

	//The rest of the pages
	currentPage := 2
	for {
		var currentUrl string
		if strings.Contains(baseUrl, "?") {
			currentUrl = fmt.Sprintf("%s&currentpage=%d", baseUrl, currentPage)
		} else {
			currentUrl = fmt.Sprintf("%s?currentpage=%d", baseUrl, currentPage)
		}
		tmpList, err = parsePageBuildingList(currentUrl)
		if err == nil {
			if len(tmpList) == 0 {
				break
			}
			letterBuildingList = append(letterBuildingList, tmpList...)
		} else {
			fmt.Sprintf("Error when parsing %s: %s", baseUrl, err.Error())
		}
		currentPage++
	}
	return letterBuildingList
}

//Given a buildings list page url it will return all the urls for the buildings on that page
//Error is possible when:
//  - Unable to retrieve the web page
//	- HTTP Response code is not 200 OK
//	- Could not parse retrieved data
func parsePageBuildingList(url string) ([]string, error) {

	var buildingsUrls []string
	resp, err := http.Get(url)

	if err != nil {
		return buildingsUrls, err
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("unexpected response from the web. HTTP code: %d", resp.StatusCode)
		return buildingsUrls, errors.New(msg)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return buildingsUrls, err
	}

	table := doc.Find("#buildings-tbl")
	rows := table.Children().First()

	rows.Children().Each(func(i int, selection *goquery.Selection) {
		//Ignoring the table heading
		if i == 0 {
			return
		}
		htmlAtag, err := selection.Children().First().Html()
		if err == nil {
			href := getStringInBetween(htmlAtag, "\"", "\"")
			bUrl := fmt.Sprintf("%s%s", BUILDING_BASE_URL, href)
			buildingsUrls = append(buildingsUrls, bUrl)
		} else {
			fmt.Printf("Could not parse a row: %s", err.Error())
		}
	})
	return buildingsUrls, nil
}

//Given a building url from the WAM page it will return the building main information
//Error is possible when:
//  - Unable to retrieve the web page
//	- HTTP Response code is not 200 OK
//	- Could not parse retrieved data
//	- The provided url is not corresponding to a correct building (building name and architect name are empty)
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
	if err != nil {
		return b, err
	}
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

	i := 3

	//Date row
	dateRow := tableHtml.Find("tr").Eq(i)
	if dateRow.Children().First().Text() == BUILDING_TABLE_DATE {
		b.date = dateRow.Children().Next().Text()
		i++
	} else {
		b.date = ""
	}

	//Style row
	styleRow := tableHtml.Find("tr").Eq(i)
	if styleRow.Children().First().Text() == BUILDING_TABLE_STYLE {
		b.style = styleRow.Find("a").Text()
		i++
	} else {
		b.style = ""
	}

	//Type row
	typeRow := tableHtml.Find("tr").Eq(i)
	if typeRow.Children().First().Text() == BUILDING_TABLE_TYPE {
		b._type = typeRow.Find("a").Text()
		i++
	} else {
		b.style = ""
	}

	//Alias row
	aliasRow := tableHtml.Find("tr").Eq(i)
	if aliasRow.Children().First().Text() == BUILDING_TABLE_ALIAS {
		b.alias = aliasRow.Children().Next().Text()
		i++
	} else {
		b.alias = ""
	}

	//Notes row
	notesRow := tableHtml.Find("tr").Eq(i)
	if notesRow.Children().First().Text() == BUILDING_TABLE_NOTES {
		b.notes = notesRow.Children().Next().Text()
	} else {
		b.notes = ""
	}

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
