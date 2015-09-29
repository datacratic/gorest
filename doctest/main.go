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
	JsonValue1 string
	JsonValue2 int
}

func RouteTestWithJson(js TestJson) string {
	return fmt.Sprintf("passed json with: %s, %d", js.JsonValue1, js.JsonValue2)
}

func main() {

	rest.AddRoute("/test", "GET", RouteTest)
	rest.AddRoute("/test/:str", "PUT", RouteTestWithStr)
	rest.AddRoute("/test/:id", "DELETE", RouteTestWithInt)
	rest.AddRoute("/test", "POST", RouteTestWithJson)
	rest.AddRoute("/test/:str/static/:id/last", "PUT", RouteTestWithStrInt)

	fmt.Println("listening")
	rest.ListenAndServe("0.0.0.0:9081", nil)

}
