package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"time"
	"github.com/fatih/color"
)

func main() {
	for {
		sleepDuration := timeUntilNextRestart()
		hours := int(sleepDuration.Hours())
		minutes := int(sleepDuration.Minutes()) % 60

		green := color.New(color.FgGreen).SprintFunc()
		fmt.Println(green("╔═════════════════════════════════════╗"))
		fmt.Println(green("║       SERVER RESTART MANAGER        ║"))
		fmt.Println(green("╚═════════════════════════════════════╝"))
		yellow := color.New(color.FgYellow).SprintfFunc()
		fmt.Println(yellow("Próximo reinício em %d horas e %d minutos...", hours, minutes))

		if sleepDuration <= 0 {
			red := color.New(color.FgRed).SprintFunc()
			fmt.Println(red("Erro: a duração do próximo reinício é inválida. Tentando novamente..."))
			time.Sleep(1 * time.Minute)
			continue
		}

		time.Sleep(sleepDuration)

		fmt.Println(yellow("Iniciando processo de reinício..."))

		stopCmd := exec.Command("taskkill", "/F", "/IM", "FXServer.exe")
		stopCmd.Run() 

		time.Sleep(10 * time.Second)

		cmd := exec.Command("cmd", "/C", `C:\Users\Administrator\Desktop\capitalcity_dev\server-data\sv_start.bat`)
		
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Println("Erro ao criar pipe para stdout:", err)
			continue
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Println("Erro ao criar pipe para stderr:", err)
			continue
		}

		err = cmd.Start()
		if err != nil {
			red := color.New(color.FgRed).SprintfFunc()
			fmt.Println(red("Erro ao iniciar o servidor: %v", err))
			continue
		}

		go printOutput(stdout, "LOG")
		go printOutput(stderr, "LOG")

		go func() {
			err = cmd.Wait()
			if err != nil {
				red := color.New(color.FgRed).SprintfFunc()
				fmt.Println(red("Servidor encerrado com erro: %v", err))
			}
		}()

		time.Sleep(1 * time.Minute)
	}
}

func printOutput(r io.Reader, prefix string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Printf("%s: %s\n", prefix, scanner.Text())
	}
}

func timeUntilNextRestart() time.Duration {
	now := time.Now()
	year, month, day := now.Date()
	loc := now.Location()

	restartTimes := []time.Time{
		time.Date(year, month, day, 7, 0, 0, 0, loc),  // 07:00
		time.Date(year, month, day, 18, 0, 0, 0, loc), // 14:10
	}

	var nextRestart time.Time
	for _, rt := range restartTimes {
		if now.Before(rt) {
			nextRestart = rt
			break
		}
	}

	if nextRestart.IsZero() {
		nextRestart = time.Date(year, month, day+1, 7, 0, 0, 0, loc)
	}

	duration := nextRestart.Sub(now)
	return duration
}
