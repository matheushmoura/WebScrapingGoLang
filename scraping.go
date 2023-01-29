package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
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

func main() {
	res, err := goquery.NewDocument("https://repositorio.seade.gov.br/group/seade-mortalidade")
	if err != nil {
		log.Fatal(err)
	}

	title := res.Find("title").Text()
	fmt.Println("Title:", title)

	res.Find("[data-format='csv']").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			linkdownload := "https://repositorio.seade.gov.br/" + link
			parts := strings.Split(link, "/")
			fileUrl := parts[len(parts)-1]
			fileUrl = "tmp/" + fileUrl + ".csv"

			err := DownloadFile(fileUrl, linkdownload)
			if err != nil {
				panic(err)
			} else {
				fmt.Println("Link:", linkdownload, " \nArquivo Baixado:", fileUrl)
				err := os.Remove(fileUrl)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("Arquivo Removido: " + fileUrl)
			}

		}
	})
}
