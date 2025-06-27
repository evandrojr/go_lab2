package main

import (
	"log"
	"os/exec"
)

func main() {
	run := func(path string) {
		cmd := exec.Command("go", "run", path)
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.Writer()
		if err := cmd.Start(); err != nil {
			log.Fatalf("Erro ao iniciar %s: %v", path, err)
		}
		go cmd.Wait()
	}

	run("cmd/a/a.go")
	run("cmd/b/b.go")

	select {} // mantém o programa rodando
}
