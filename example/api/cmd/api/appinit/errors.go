package appinit

import "log"

func fatalError(err error, msgs ...any) {
	if err != nil {
		log.Fatalln(append(msgs, err)...)
	}
}
