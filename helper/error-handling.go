package helper

func HardCheck(e error) {
	if e != nil {
		panic(e)
	}
}