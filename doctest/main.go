package main

import (
	"github.com/datacratic/gorest/rest"

	"fmt"
)

func RouteTest() string {
	return "simple GET"
}

func RouteTestWithStr(str string) string {
	return fmt.Sprintf("passed string: %s", str)
}

func RouteTestWithStrInt(str string, id int) string {
	return fmt.Sprintf("passed string: %s, %d", str, id)
}

func RouteTestWithInt(id int) string {
	return fmt.Sprintf("passed int: %d", id)
}

type TestJson struct {
	JsonValue1 string `json:"jsonValue1"`
	JsonValue2 int    `json:"jsonValue2:`
}

func RouteTestWithJson(js TestJson) string {
	return fmt.Sprintf("passed json with: %s, %d", js.JsonValue1, js.JsonValue2)
}

type TestJson2 struct {
	J  *TestJson
	S  *string
	SS []string
	M  map[string]int
}

func RouteTestWithJson2(js TestJson2) string {
	return fmt.Sprintf("passed json with: %s, %d, %s", js.J.JsonValue1, js.J.JsonValue2, js.S)
}

func main() {

	rest.AddRoute("/test", "GET", RouteTest)
	rest.AddRoute("/test/:str", "PUT", RouteTestWithStr)
	rest.AddRoute("/test/:id", "DELETE", RouteTestWithInt)
	rest.AddRoute("/test", "POST", RouteTestWithJson)
	rest.AddRoute("/test2", "POST", RouteTestWithJson2)
	rest.AddRoute("/test/:str/static/:id/last", "PUT", RouteTestWithStrInt)

	fmt.Println("listening")
	rest.ListenAndServe("0.0.0.0:9081", nil)

}
