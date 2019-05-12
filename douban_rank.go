package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

const (
	MovieRankingTop250    = "top250" 		// ranking table "Top250"
	UsboxRankingTop10     = "us_box" 		// ranking table "us_box"
	MovieRankingTop250Num = 250      		// Number of movie items in the Top250 rank.
	UsboxRankingTop10Num  = 10             	// Number of movie items in the us_box rank.
)

type RankAttr struct {
	URL   string
	total int
	resultType reflect.Type
}

func (attr RankAttr) Total() int {
	return attr.total
}

var UrlMap = map[string]RankAttr{
	MovieRankingTop250: 
		{"http://api.douban.com/v2/movie/top250?count=%d&start=%d", MovieRankingTop250Num, reflect.TypeOf((*MovieRanking250)(nil)).Elem()},
	UsboxRankingTop10:	
		{"http://api.douban.com/v2/movie/us_box?count=%d&start=%d", UsboxRankingTop10Num, reflect.TypeOf((*UsboxRanking)(nil)).Elem()},
}

type Artist struct {
	Link    string            `json:"alt"`
	Avatars map[string]string `json:"avatars"`
	Name    string            `json:"name"`
	ID      string            `json:"id"`
}

func (self *Artist) String() string {
	return self.Name
}

type Rate struct {
	Max     int     `json:"max"`
	Average float32 `json:"average"`
	Stars   string  `json:"stars"`
	Min     int     `json:"min"`
}

type MovieItem struct {
	Rating       Rate              `json:"rating"`
	Genres       []string          `json:"genres"`
	Title        string            `json:"title"`
	Casts        []Artist          `json:"casts"`
	CollectCount int               `json:"collect_count"`
	OriginName   string            `json:"original_title"`
	SubType      string            `json:"subtype"`
	Directors    []Artist          `json:"directors"`
	Year         string            `json:"year"`
	Imags        map[string]string `json:"images"`
	URL          string            `json:"alt"`
	ID           string            `json:"id"`
}

func (self *MovieItem) GetAverageScore() float32 {
	return self.Rating.Average
}

func (self *MovieItem) GetGenres() string {
	var buffer bytes.Buffer
	strLen := 0

	if self.Genres == nil {
		return ""
	}

	for _, val := range self.Genres {
		l, _ := buffer.WriteString(val)
		buffer.WriteString(",")
		strLen += l + 1
	}
	buffer.Truncate(strLen - 1)

	return buffer.String()
}

func (self *MovieItem) GetCasts() string {
	if self.Casts == nil {
		return ""
	}
	
	return getStarrings(self.Casts)
}

func (self *MovieItem) GetDirectors() string {
	if self.Directors == nil {
		return ""
	}

	return getStarrings(self.Directors)
}

func getStarrings(artists []Artist) string {
	var buffer bytes.Buffer
	strLen := 0

	for _, val := range artists {
		l, _ := buffer.WriteString(val.String())
		buffer.WriteString(",")
		strLen += l + 1
	}
	if strLen == 0 {
		return ""
	} else {
		buffer.Truncate(strLen - 1)
		return buffer.String()
	}
}

type UsboxItem struct {
	Box     int       `json:"box"`
	Isnew   bool      `json:"new"`
	Rank    int       `json:"rank"`
	Subject MovieItem `json:"subject"`
}

// The box rank of USA
type UsboxRanking struct {
	RankTitle string      `json:"title"`
	RankDate  string      `json:"date"`
	Subjects  []UsboxItem `json:"subjects"`
}

func (usboxRanking *UsboxRanking) extractMovieItems() ([]MovieItem) {
	result := make([]MovieItem, 0, len(usboxRanking.Subjects))

	for _, v := range usboxRanking.Subjects {
		result = append(result, v.Subject)
	}

	return result
}

// The Douban movie250
type MovieRanking250 struct {
	Count    int         `json:"count"`
	Start    int         `json:"start"`
	Total    int         `json:"total"`
	Subjects []MovieItem `json:"subjects"`
}

func arrayToString(arr []fmt.Stringer) string {
	var buffer bytes.Buffer
	strLen := 0

	for _, val := range arr {
		l, _ := buffer.WriteString(val.String())
		buffer.WriteString(",")
		strLen += l + 1
	}
	buffer.Truncate(strLen - 1)

	return buffer.String()
}

func RetrieveRankingData(rankName string, length int) ([]MovieItem, error) {
	var output []MovieItem
	var attr RankAttr
	var ok bool

	if attr, ok = UrlMap[rankName]; !ok {
		return nil, fmt.Errorf("specified rank name\"%s\" does not exist", rankName)
	}

	var routineNum int
	if attr.total%length > 0 {
		routineNum = attr.total/length + 1
	} else {
		routineNum = attr.total / length
	}
	resultChannel := make(chan []MovieItem, routineNum)
	errorChannel := make(chan error, routineNum)
	resultSet := make([]MovieItem, 0, attr.total)

	for i := 0; i < attr.total; i += length {
		go func(offset int, length int) {
			client := &http.Client{}
			ranking := reflect.New(attr.resultType)

			apiURL := fmt.Sprintf(attr.URL, length, offset)
			fmt.Printf("[DEBUG]Requesting %s\n", apiURL)

			req, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				errorChannel <- err
				return
			}

			req.Header.Set("Accept-Charset", "utf-8;q=0.7,*;q=0.3")
			req.Header.Set("User-Agent", "chrome 100")

			resp, reqError := client.Do(req)
			if reqError != nil {
				errorChannel <- reqError
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				bodyContent, readErr := ioutil.ReadAll(resp.Body)
				if readErr != nil {
					errorChannel <- readErr
					return
				}
				unmarshalErr := json.Unmarshal(bodyContent, ranking.Interface())
				if unmarshalErr != nil {
					errorChannel <- unmarshalErr
					return
				}

				if strings.Contains(attr.resultType.Name(), "MovieRanking250") {
					if r, ok := ranking.Interface().(*MovieRanking250); ok {
						resultChannel <- r.Subjects
					} else {
						errorChannel <- fmt.Errorf("Convertion to MovieRanking250 failed")
					}
				} else if (strings.Contains(attr.resultType.Name(), "UsboxRanking")) {
					if r, ok := ranking.Interface().(*UsboxRanking); ok {
						resultChannel <- r.extractMovieItems()
					} else {
						errorChannel <- fmt.Errorf("Convertion to UsboxRanking failed")
					}
				} else {
					errorChannel <- fmt.Errorf("Unknown resultType: %s", attr.resultType.Name())
				}
				//resultChannel <- ranking.Subjects
				return
			}
			errorChannel <- fmt.Errorf("Bad status code: %d", resp.StatusCode)
			return
		}(i, length)

	}

	for i := 0; i < routineNum; i++ {
		select {
		case output = <-resultChannel:
			resultSet = append(resultSet, output...)
			//fmt.Printf("[DEBUG] %d items retrieved. The total result length: %d.\n", len(output), len(resultSet))
		case e := <-errorChannel:
			return nil, e
		}
	}

	return resultSet, nil
}
