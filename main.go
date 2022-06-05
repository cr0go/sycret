package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
)

type params struct {
	ns       string
	tag      string
	atr      string
	template string
}

type person struct {
	Cardnumber string
	Surname    string
	Name       string
	Secondname string
}

type response struct {
	Result            int
	ResultDescription string
	ResultData        string
}

type rquery struct {
	templateURL string
	RecordID    string
}

func (p *person) String() string {
	return fmt.Sprintf("%s %s %s %s", p.Cardnumber, p.Surname, p.Name, p.Secondname)
}

//regExp returns regular expression according to params
func regExp(p params) string {
	return fmt.Sprintf(p.template, p.ns, p.tag, p.atr)
}

var (
	rqu     = rquery{}
	dataUrl = "https://sycret.ru/service/apigendoc/apigendoc"
	p       = params{ns: "ns1", tag: "text", atr: "field", template: "//%s:%s/@%s"}
)

func main() {

	handler()

	// {
	// 	"URLTemplate": "https://sycret.ru/service/apigendoc/forma_025u.doc",
	// 	"RecordID": 30
	// }

	// {
	// 	"URLWord": "your_url\2022-05-26 14-12-04.doc"
	// }
}

func initial(w http.ResponseWriter) {

	list, doc, err := getTemplateFields(rqu, p)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if err := getPersonData(dataUrl, p, list); err != nil {
		log.Fatalf("Error: %v", err)
	}

	saveToDoc(list, doc, w)
}

func saveToDoc(data map[string]person, doc *xmlquery.Node, w http.ResponseWriter) {
	fname := fmt.Sprintf(`%s.doc`, time.Now().Format("2006-01-02 15-04-05"))
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Error file creating: %q", err)
	}
	for k, v := range data {
		re := fmt.Sprintf("//*[@field=\"%s\"]/*/*[2]", k)
		node := xmlquery.FindOne(doc, re)
		node.FirstChild.Data = v.String()
	}
	res := doc.OutputXML(true)
	_, err = f.WriteString(res)
	if err != nil {
		log.Fatalf("Error write file: %q", err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalf("Error close file : %q", err)
	}

	filepath := fmt.Sprintf("%s\\%s", "localhost:8000", fname)

	fmt.Println(filepath)
	json.NewEncoder(w).Encode(map[string]string{"URLWord": filepath})
	// {
	// "URLWord": "your_url\2022-05-26 14-12-04.doc"
	// }
}

//getPersonData fill map with values
func getPersonData(uurl string, p params, m map[string]person) error {
	u, err := url.Parse(uurl)
	if err != nil {
		return err
	}
	for k := range m {
		q := u.Query()
		q.Add(p.tag, k)
		u.RawQuery = q.Encode()
		b, err := getData(u.String())
		if err != nil {
			return err
		}
		r := unmarshalResponse(b)
		p, err := r.getPerson()
		if err != nil {
			return err
		}
		m[k] = p
	}
	return nil
}

//getTemplateFields returns map with found regexp results
func getTemplateFields(rqu rquery, p params) (map[string]person, *xmlquery.Node, error) {
	re := regExp(p)
	b, err := getData(rqu.templateURL)
	if err != nil {
		return nil, nil, err
	}
	doc, err := xmlquery.Parse(strings.NewReader(string(b)))
	if err != nil {
		return nil, nil, err
	}
	list := xmlquery.Find(doc, re)
	res := make(map[string]person)
	for _, v := range list {
		res[v.InnerText()] = person{}
	}
	return res, doc, nil
}

//getData does request with specific params
//and returns response's body []byte
func getData(u string) ([]byte, error) {
	client := &http.Client{}
	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("User-Agent", "whatever")
	r.Method = "GET"
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

//unmarshalResponse returns response{}
func unmarshalResponse(b []byte) response {
	r := response{}
	err := json.Unmarshal(b, &r)
	if err != nil {
		log.Fatalln(err)
	}
	return r
}

//getPerson creates new person{} from response{}
func (r *response) getPerson() (person, error) {
	p := person{}
	if len(r.ResultData) == 0 {
		return p, errors.New("response has no data")
	}
	data := strings.Fields(r.ResultData)
	p.Surname = data[0]
	p.Name = data[1]
	p.Secondname = data[2]
	p.Cardnumber = rqu.RecordID
	return p, nil
}
