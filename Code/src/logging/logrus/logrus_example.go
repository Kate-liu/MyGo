package logrus

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Print("call Print: line1")
	log.Println("call Println: line2")
}
