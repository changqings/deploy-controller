package main

import (
	local_mgr "pod-controller/manager"
)

func main() {
	err := local_mgr.RunManager()
	if err != nil {
		panic(err)
	}
}
