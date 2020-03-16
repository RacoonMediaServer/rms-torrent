package utils

import "io/ioutil"

func DumpPage(name string, content []byte) {
	if err := ioutil.WriteFile(name+".html", content, 0644); err != nil {
		panic(err)
	}
}
