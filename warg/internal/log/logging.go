package log

import "fmt"

func Log(s string) {
	fmt.Println(s)
}

func Stdout(s string) {

}

func Debug(s string) {

}

func Warn(s string) {

}

func Error(s string) {
	Log(s)
}


