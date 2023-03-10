package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-gota/gota/dataframe"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"gopkg.in/mgo.v2"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func FileToDataframe(filepath string) (dataframe.DataFrame, error) {
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
		return dataframe.DataFrame{}, err
	}
	defer f.Close()

	reader := csv.NewReader(transform.NewReader(f, charmap.Windows1252.NewDecoder()))
	reader.Comma = ';'
	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return dataframe.DataFrame{}, err
	}
	df := dataframe.LoadRecords(lines)
	return df, nil
}

func FileToMongo(filepath string) {
	session, err := mgo.Dial("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer session.Close()

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Erro ao abrir CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Erro ao ler CSV: %v", err)
	}

	collection := session.DB("Scraping").C("Data")

	docs := make([]interface{}, len(data))
	for i, record := range data {
		doc := make(map[string]interface{})
		for j, header := range data[0] {
			doc[header] = record[j]
		}
		docs[i] = doc
	}

	err = collection.Insert(docs...)
	if err != nil {
		log.Fatalf("Erro ao inserir no MongoDB: %v", err)
	}

	fmt.Println("Exportado para o MongoDB com sucesso!")
}

func DirectToDownload(linkDownload string) {
	resDownload, erroDownload := goquery.NewDocument(linkDownload)
	if erroDownload != nil {
		log.Fatal(erroDownload)
	}
	downloadButton, existsButton := resDownload.Find("a:has(i[class='fa fa-arrow-circle-o-down'])").Attr("href")
	if existsButton {
		parts := strings.Split(downloadButton, "/")
		fileUrl := parts[len(parts)-1]
		fileUrl = "tmp/" + fileUrl

		err := DownloadFile(fileUrl, downloadButton)
		if err != nil {
			panic(err)
		} else {
			fmt.Println("Link:", downloadButton, " \nArquivo Baixado:", fileUrl)

			df, erro := FileToDataframe(fileUrl)
			if erro != nil {
				panic(err)
			} else {
				fmt.Println(df)
			}

			FileToMongo(fileUrl)

			err := os.Remove(fileUrl)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Arquivo Removido: " + fileUrl)
		}
	}
}

func main() {
	res, err := goquery.NewDocument("https://repositorio.seade.gov.br/dataset/municipios")
	if err != nil {
		log.Fatal(err)
	}

	title := res.Find("title").Text()
	fmt.Println("Title:", title)

	res.Find("ul.resource-list > li.resource-item > a:has(span[data-format='csv'])").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			linkdownload := "https://repositorio.seade.gov.br" + link
			DirectToDownload(linkdownload)
		}
	})
}
