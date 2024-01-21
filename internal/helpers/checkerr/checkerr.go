package checkerr

import "log"

func CheckErr(message string, e error) {
	if e != nil {
		log.Fatalf(message, e)
	}
}
