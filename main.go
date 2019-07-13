package main

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	//itemList := getItemList()

	i := &Item{Id: "1121212", Url: "https://simplewine.ru/catalog/product/chateau_d_armailhac_2011_2/"}
	i.parse()

}

func getItemList() []*Item {
	return nil
}

type Item struct {
	Id  string
	Url string

	body []byte

	Facts []string

	Descriptions map[string]string

	Params map[string]string

	ImageUrl string

	Rates map[string]int
}

func (i *Item) parse() error {
	err := i.downloadBody()
	if err != nil {
		return err
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
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	i.body = data

	return nil
}

func (i *Item) parseFacts(document *goquery.Document) error {
	sel := document.Find(".product-facts__link")
	sel.Each(i.eachFact)
	return nil
}

func (i *Item) eachFact(n int, selection *goquery.Selection) {
	data := selection.Nodes[0].FirstChild.Data
	if i.Facts == nil {
		i.Facts = make([]string, 0)
	}
	i.Facts = append(i.Facts, san(data))
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
