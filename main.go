package main

import (
	_ "embed"
	"fmt"
	"github.com/ulstu-schedule/parser/schedule"
)

func main() {
	mapa := schedule.GetGroups()
	fmt.Print(mapa)
	//
	//data, err := easyjson.Marshal(groupSchedule)
	//
	//fmt.Printf("%s", data)
	//
	//copySchedule := &types.Schedule{}
	//
	//err = easyjson.Unmarshal(data, copySchedule)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Printf("\n%s", copySchedule.Weeks[0].Days[0])
}

//var incorrectDateError *types.IncorrectDateError
//
//func main() {
//	date, err := schedule.GetTextDayGroupScheduleByDate("ПИбд-21", "32.10")
//	if err != nil {
//		switch {
//		case errors.As(err, &incorrectDateError):
//			fmt.Println("Ошибка в дате!")
//			fmt.Println("Дата-то вот: " + incorrectDateError.Date)
//		default:
//			fmt.Println("Это что за ошибка?")
//		}
//	}
//	fmt.Println(date)
//}

//func getBody(id string) (*http.Response, error) {
//	response, err := get(fmt.Sprintf("https://vk.com/%s", id))
//	if err != nil {
//		return nil, fmt.Errorf("error getting body of id=%s: %w", id, err)
//	}
//	return response, err
//}
//
//func get(url string) (*http.Response, error) {
//	response, err := http.Get(url)
//	if err != nil {
//		return nil, err
//	}
//	if response.StatusCode != 200 {
//		return nil, &types.StatusCodeError{StatusCode: response.StatusCode, StatusText: response.Status}
//	}
//	return response, nil
//}
