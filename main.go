package main

func main() {
	sol := newSolution()
	err := sol.run()
	if err != nil {
		panic(err)
	}
}
