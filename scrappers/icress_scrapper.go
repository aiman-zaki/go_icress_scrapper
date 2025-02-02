package scrapper

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"context"
	"log"

	models "github.com/aiman-zaki/go_central_scrapper/models"

	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getClient() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://admin:password@localhost:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func removeSpace(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

//FetchFaculties : to fetch all faculties
func FetchFaculties() []models.Faculty {
	var faculties []models.Faculty
	const URL = "http://icress.uitm.edu.my/jadual/jadual/jadual.asp"
	c := colly.NewCollector(
		colly.AllowedDomains("icress.uitm.edu.my"),
	)
	c.OnHTML("option", func(e *colly.HTMLElement) {
		temp := models.Faculty{}
		tempString := strings.Split(e.Text, "-")
		temp.Code = tempString[0]
		temp.Name = tempString[1]
		faculties = append(faculties, temp)
		println(faculties)
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	c.Visit(URL)
	return faculties
}

//FetchCourses := fetch all courses from a faculty
func FetchCourses(f models.Faculty) models.Faculty {
	var courses []models.Course
	var URL = "http://icress.uitm.edu.my/jadual/" + f.Code + "/" + f.Code + ".html"
	c := colly.NewCollector(
		colly.AllowedDomains("icress.uitm.edu.my"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*icress.*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		course := models.Course{}
		course.Code = e.Text
		course.Timetables = FetchTimetables(f.Code, course.Code)
		courses = append(courses, course)
		f.Courses = courses
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	c.Visit(URL)

	c.Wait()
	return f

}

//FetchTimetables : fetch all timetables from a course
func FetchTimetables(faculty string, course string) []models.Timetable {
	var URL = "http://icress.uitm.edu.my/jadual/" + faculty + "/" + course + ".html"
	var timetables []models.Timetable

	c := colly.NewCollector(
		colly.AllowedDomains("icress.uitm.edu.my"),
	)

	c.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		temp := models.Timetable{}
		e.ForEach("td", func(index int, el *colly.HTMLElement) {
			trimpedText := removeSpace(el.Text)
			switch index {
			case 0:
				temp.Group = trimpedText
			case 1:
				temp.Start = trimpedText
			case 2:
				temp.End = trimpedText
			case 3:
				temp.Day = trimpedText
			case 4:
				temp.Mode = trimpedText
			case 5:
				temp.Status = trimpedText
			case 6:
				temp.Room = trimpedText
			}

		})
		fmt.Printf("%+v\n", temp)
		timetables = append(timetables, temp)

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	c.Visit(URL)
	return timetables

}
