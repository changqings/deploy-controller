package main

import local_manager "deploy-controller/manager"

func main() {
	err := local_manager.RunManager()
	if err != nil {
		panic(err)
	}
}
