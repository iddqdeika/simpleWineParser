package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"gopkg.in/cheggaaa/pb.v1"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {

	itemList, err := getItemList()
	if err != nil {
		panic(err)
	}

	bar := pb.StartNew(len(itemList))
	lock := sync.Mutex{}
	wg := sync.WaitGroup{}
	ch := make(chan struct{}, 20)

	for _, i := range itemList {
		wg.Add(1)
		go func(item *Item) {
			defer wg.Done()
			ch <- struct{}{}
			err := item.parse()
			lock.Lock()
			item.Error = err
			bar.Increment()
			lock.Unlock()

			<-ch
		}(i)

	}
	wg.Wait()

	bar.Finish()

	err = writeResults(itemList)
	if err != nil {
		fmt.Printf("error during result writing: %v", err)
	}

}

func getItemList() ([]*Item, error) {
	file, err := excelize.OpenFile("zakaz.xlsx")

	if err != nil {
		return nil, err
	}

	result := make([]*Item, 0)
	met := make(map[string]struct{})
	var i int = 1
	for {
		i++
		val, err := file.GetCellValue("Лист1", "E"+strconv.Itoa(i))
		if err != nil {
			return nil, err
		}
		if len(val) == 0 {
			break
		}
		id, err := file.GetCellValue("Лист1", "A"+strconv.Itoa(i))
		if err != nil {
			return nil, err
		}
		name, err := file.GetCellValue("Лист1", "C"+strconv.Itoa(i))
		if err != nil {
			return nil, err
		}

		if _, ok := met[id]; !ok {
			result = append(result, &Item{Id: id, Url: val, Name: name})
			met[id] = struct{}{}
		}
	}

	return result, nil
}

func writeResults(items []*Item) error {

	file := excelize.NewFile()
	t := Table{}
	for _, i := range items {
		if i.Error == nil {
			for k, v := range i.Rates {
				t.AddRow()
				t.SetCellValue("id", i.Id)
				t.SetCellValue("param", k)
				t.SetCellValue("value", strconv.Itoa(v))
				t.SetCellValue("name", i.Name)

			}

			for k, v := range i.Params {
				t.AddRow()
				t.SetCellValue("id", i.Id)
				t.SetCellValue("param", k)
				t.SetCellValue("value", v)
				t.SetCellValue("name", i.Name)
			}

			for k, v := range i.Descriptions {
				t.AddRow()
				t.SetCellValue("id", i.Id)
				t.SetCellValue("param", k)
				t.SetCellValue("value", v)
				t.SetCellValue("name", i.Name)
			}

			for n, v := range i.Facts {
				t.AddRow()
				t.SetCellValue("id", i.Id)
				t.SetCellValue("param", "fact "+strconv.Itoa(n))
				t.SetCellValue("value", v)
				t.SetCellValue("name", i.Name)
			}
		}
	}

	writeTableToXlsx(file, "sheet1", t)

	err := file.SaveAs("test.xlsx")

	return err
}

//write Table object content to given xlsx file into sheet with given name
func writeTableToXlsx(xlsx *excelize.File, sheetname string, table Table) {
	fmt.Println("\t" + "writing sheet \"" + sheetname + "\"...")
	xlsx.NewSheet(sheetname)
	for k, v := range table.Columns {
		columnname := getColumnName(v)
		xlsx.SetCellValue(sheetname, columnname+"1", k)
	}
	var i int

	bar := pb.StartNew(len(table.Rows))
	for k, v := range table.Rows {
		i++
		bar.Increment()
		rowname := strconv.Itoa(k + 2)

		//xlsx.SetSheetRow(sheetname, "A" + rowname,&sl)
		for kk, vv := range v.Cells {
			columnname := getColumnName(kk)
			xlsx.SetCellValue(sheetname, columnname+rowname, vv)
		}

	}
	bar.Finish()
}

//write Table object content to given xlsx file into sheet with given name
func writeTableToCsv(filename string, table Table) {
	fmt.Println("\t" + "writing file \"" + filename + "\"...")

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		file, err = os.Create(filename)
	}
	file.Close()

	fileHandle, _ := os.OpenFile(filename, os.O_APPEND, 0666)
	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()

	maxcol := 0
	colMap := make(map[int]string)
	for colName, colNum := range table.Columns {
		colMap[colNum] = colName
		if maxcol < colNum {
			maxcol = colNum
		}
	}
	data := ""
	for i := 0; i < maxcol; i++ {
		data += colMap[i] + "\t"
	}
	data += "\r\n"
	fmt.Fprint(writer, data)

	bar := pb.StartNew(len(table.Rows))
	for _, row := range table.Rows {
		maxcol := 0
		for _, colNum := range table.Columns {
			if _, ok := row.Cells[colNum]; ok {
				if maxcol < colNum {
					maxcol = colNum
				}
			}
		}
		data := ""
		for i := 0; i < maxcol; i++ {
			data += row.Cells[i] + "\t"
		}
		data += "\r\n"
		fmt.Fprint(writer, data)
		//file.Sync()
		bar.Increment()
	}
	bar.Finish()
}

//function to get column name by column number. supports up to 17526 columns
func getColumnName(v int) string {
	var columnname string
	if v <= 17526 {
		alfabet := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
		columnnum := v
		if columnnum >= len(alfabet) {
			if columnnum >= len(alfabet)*len(alfabet) {
				first := (columnnum - (columnnum % (len(alfabet) * len(alfabet)))) / (len(alfabet) * len(alfabet))
				last := (columnnum - first) % len(alfabet)
				mid := (columnnum - (first * len(alfabet) * len(alfabet)) - last) / len(alfabet)
				columnname = alfabet[first] + alfabet[mid] + alfabet[last]
			} else {
				last := alfabet[columnnum%len(alfabet)]
				first := alfabet[columnnum/len(alfabet)-1]
				columnname = first + last
			}
		} else {
			first := alfabet[columnnum]
			columnname = first
		}
	} else {
		columnname = getColumnName(17526)
	}
	return columnname
}

type Item struct {
	Id    string
	Ready bool
	Error error
	Name  string
	Url   string

	body []byte

	Facts []string

	Descriptions map[string]string

	Params map[string]string

	ImageUrl string

	Rates map[string]int
}

func (i *Item) parse() error {
	for {
		err := i.downloadBody()
		if err != nil {
			return err
		} else {
			break
		}
	}

	node, err := html.Parse(bytes.NewReader(i.body))
	if err != nil {
		return err
	}

	doc := goquery.NewDocumentFromNode(node)

	err = i.parseFacts(doc)
	if err != nil {
		return err
	}

	err = i.parseDescriptions(doc)
	if err != nil {
		return err
	}

	err = i.parseParams(doc)
	if err != nil {
		return nil
	}

	err = i.parseRates(doc)
	if err != nil {
		return nil
	}
	return err
}

func (i *Item) downloadBody() error {
	resp, err := http.Get(i.Url)
	if err != nil {
		return err
	}

	if resp.Request.URL.String() != i.Url {
		return errors.New("redirect")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	i.body = data

	return nil
}

func (i *Item) parseFacts(document *goquery.Document) error {
	sel := document.Find(".product-facts__value")
	sel.Each(i.eachFact)
	return nil
}

func (i *Item) eachFact(n int, selection *goquery.Selection) {
	sel := selection.Find(".product-facts__link")

	result := ""

	sel.Each(func(n int, selection *goquery.Selection) {
		data := selection.Nodes[0].FirstChild.Data
		if len(result) > 0 {
			result += ","
		}
		result += san(data)
	})

	if i.Facts == nil {
		i.Facts = make([]string, 0)
	}
	i.Facts = append(i.Facts, result)
}

func (i *Item) parseDescriptions(document *goquery.Document) error {
	sel := document.Find(".characteristics-description__item")
	sel.Each(i.eachDescriptionItem)

	return nil
}

func (i *Item) eachDescriptionItem(n int, selection *goquery.Selection) {
	title := selection.Find(".characteristics-description__item-title").Nodes[0].FirstChild.Data
	desc := selection.Find(".characteristics-description__item-text").Nodes[0].FirstChild.Data

	if i.Descriptions == nil {
		i.Descriptions = make(map[string]string)
	}

	i.Descriptions[san(title)] = san(desc)
}

func (i *Item) parseParams(document *goquery.Document) error {
	sel := document.Find(".characteristics-params__item")
	sel.Each(i.eachParam)

	return nil
}

func (i *Item) eachParam(n int, selection *goquery.Selection) {
	title := selection.Find(".characteristics-params__title").Nodes[0].FirstChild.Data
	valueSelection := selection.Find(".characteristics-params__value")
	as := valueSelection.Find("a")
	var value string
	if as.Nodes != nil && len(as.Nodes) > 0 {
		value = as.Nodes[0].FirstChild.Data
	} else {
		value = valueSelection.Nodes[0].FirstChild.Data
	}

	if i.Params == nil {
		i.Params = make(map[string]string)
	}

	i.Params[san(title)] = san(value)
}

func (i *Item) parseRates(document *goquery.Document) error {
	sel := document.Find(".product-info__meta-item")
	sel.Each(i.eachRate)

	return nil
}

func (i *Item) eachRate(n int, selection *goquery.Selection) {
	if strings.Contains(selection.Nodes[0].Attr[0].Val, "info__meta-item_rating") {
		return
	}
	rateName := san(selection.Nodes[0].FirstChild.Data)

	vals := selection.Find(".product-info__meta-item-value")
	if vals.Nodes == nil || len(vals.Nodes) <= 0 {
		return
	}
	rateValue, err := strconv.Atoi(san(vals.Nodes[0].FirstChild.Data))
	if err != nil {
		return
	}

	if i.Rates == nil {
		i.Rates = make(map[string]int)
	}
	i.Rates[rateName] = rateValue
}

func san(text string) string {
	return strings.Trim(text, " \n")
}
