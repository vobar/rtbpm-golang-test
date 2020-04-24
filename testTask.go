package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
)
//import "log"
//import "time"
type resultsTest struct {
	path string
	count int
}
func main() {

	//сборщик результатов
	resChan := make(chan resultsTest)

	//флаг выполнения все горутин
	done := make(chan struct{})

	go func(results <-chan resultsTest, done chan<- struct{}) {
		var totalCount int
		for result := range results {
			fmt.Printf("Count for %s: %d\n", result.path, result.count)
			totalCount += result.count
		}
		fmt.Printf("Total: %d\n", totalCount)
		done <- struct{}{}
	}(resChan, done)

	//read stdin
	scanner := bufio.NewScanner(os.Stdin)

	var wg sync.WaitGroup
	//ограничитель кол-ва горутинок для чтения линков
	goroutinesChan := make(chan struct{}, 5)
	//read strings from stdin
	for scanner.Scan() {
		link := scanner.Text()
		goroutinesChan <- struct{}{}
		wg.Add(1)
		go func() {
			//GET запрос
			resp, err := http.Get(link)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			aORb := regexp.MustCompile("Go")
			matches := aORb.FindAllStringIndex(string(body), -1)

			//закидываем данные в канал результата
			resChan <- resultsTest{
				link,
				len(matches),
			}
			//уменьшаем счётчик горутинок
			<-goroutinesChan
			wg.Done()
		}()
	}

	wg.Wait()
	close(goroutinesChan)
	close(resChan)
	//ждём завершения горутин
	<-done
}
