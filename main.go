package main

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
)

func main() {
	//itemList := getItemList()

	i := &Item{Id:"1121212", Url:"https://simplewine.ru/catalog/product/chateau_d_armailhac_2011_2/"}
	i.parse()


}

func getItemList() []*Item {
	return nil
}

type Item struct {
	Id				string
	Url				string

	body			[]byte

	Facts			[]string

	Descriptions	map[string]string

	Params			[]Param

	ImageUrl		string

	Rates			map[string]int
}

type Param struct {
	Name	string
	Value	string
}

func (i *Item) parse() error{
	err := i.downloadBody()
	if err != nil{
		return err
	}
	node, err := html.Parse(bytes.NewReader(i.body))
	if err != nil{
		return err
	}

	doc := goquery.NewDocumentFromNode(node)

	err = i.parseFacts(doc)
	if err != nil{
		return err
	}


	return err
}

func (i *Item) downloadBody() error{
	resp, err := http.Get(i.Url)
	if err != nil{
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return err
	}
	i.body = data

	return nil
}

func (i *Item) parseFacts(document *goquery.Document) error{
	sel := document.Find(".product-facts__link")
	sel.Each(i.eachFact)
	return nil
}

func (i *Item) eachFact(n int, selection *goquery.Selection) {
	data := selection.Nodes[0].FirstChild.Data
	if i.Facts == nil{
		i.Facts = make([]string,0)
	}
	i.Facts = append(i.Facts, data)
}

