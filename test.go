package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {

	// cronosd config set client chain-id cronos_777-1
	// cronosd config set client keyring-backend test
	// cronosd config set client broadcast-mode sync
	// cronosd init test --chain-id cronos_777-1 --overwrite
	// cronosd keys add validator --keyring-backend test
	// cronosd genesis add-genesis-account validator 1500000000000000000000stake,1500000000000000000000basetcro --keyring-backend test
	// cronosd genesis gentx validator 1000000000000000000000stake --chain-id cronos_777-1 --moniker="validator" --min-self-delegation="1000000" --ip="127.0.0.1" --keyring-backend test
	// cronosd genesis collect-gentxs

	// genesis.json: max_gas: "81500000", "cronos_admin": "crc1ujg55plwu3x4kl95606yjk27frjvssqtkn59hp", "evm_denom": "basetcro"
	// app.toml: pruning = "everything", minimum-gas-prices = "0basetcro"
	// config.toml: timeout_commit = "1.5s"

	errMsg := "Value missing for key"
	home := "/Users/mavis/.cronos"
	for i := 0; i < 1; i++ {
		err := os.RemoveAll(home)
		if err != nil {
			fmt.Println("err when remove cronosd:", err)
			continue
		}

		err = os.Mkdir(home, 0777)
		if err != nil {
			fmt.Println("err when make cronosd:", err)
			continue
		}

		cmd := exec.Command("cp", "-R", "/Users/mavis/Desktop/cronos/", home)
		fmt.Println("mm-cp")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("err when cp:", err)
			continue
		}

		l := "cmd.log"
		if _, err := os.Stat(l); err == nil {
			err = os.Remove(l)
			if err != nil {
				panic(err)
			}
		}

		file, err := os.OpenFile(l, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("Error when open file:", err)
			continue
		}

		fmt.Println("mm-start1")
		cronosdCmd := exec.Command("cronosd", "start")
		go cronosdCmd.Run()
		time.Sleep(time.Second * 3)
		fmt.Println("mm-kill1")
		cronosdCmd.Process.Kill()
		cronosdCmd.Process.Wait()

		for i := 0; i < 20; i++ {
			fmt.Println("mm-start2")
			cronosdCmd = exec.Command("cronosd", "start")
			cronosdCmd.Stdout = file
			cronosdCmd.Stderr = os.Stderr
			go cronosdCmd.Run()
			time.Sleep(time.Second * 3)

			// cmd := exec.Command("cronosd", "tx", "e2ee", "register-encryption-key", "age1ujga6nskz2vs9st9f2rar66c6s3vprrpzfe06ukuc8ttdazdecpqc4qfuh", "--from", "validator", "--keyring-backend", "test", "--gas", "auto", "--gas-adjustment", "1.5", "--gas-prices", "100000000000basetcro", "--chain-id", "cronos_777-1", "-y")
			cmd := exec.Command("cronosd", "tx", "cronos", "update-permissions", "crc1ujg55plwu3x4kl95606yjk27frjvssqtkn59hp", "3", "--from", "validator", "--keyring-backend", "test", "--gas", "auto", "--gas-adjustment", "1.5", "--gas-prices", "100000000000basetcro", "--chain-id", "cronos_777-1", "-y")
			fmt.Println("mm-tx", i)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println("mm-tx-err", i, err)

				fmt.Println("killing proc")
				cronosdCmd.Process.Kill()
				fmt.Println("killed proc")
				cronosdCmd.Process.Wait()
				fmt.Println("proc end")
				if i > 2 {
					break
				}
				continue
			}

			errCount := 0
			maxErrCount := 2
			for {
				time.Sleep(time.Second * 1)
				// cmd = exec.Command("cronosd", "q", "e2ee", "key", "crc1ujg55plwu3x4kl95606yjk27frjvssqtkn59hp")
				cmd = exec.Command("cronosd", "q", "cronos", "permissions", "crc1ujg55plwu3x4kl95606yjk27frjvssqtkn59hp")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					fmt.Println("mm-query-err", errCount)
					if errCount += 1; errCount < maxErrCount {
						continue
					}
				} else {
					fmt.Println("mm-query-success")
					time.Sleep(time.Second * 2)

					fmt.Println("killing proc")
					cronosdCmd.Process.Kill()
					fmt.Println("killed proc")
					cronosdCmd.Process.Wait()
					fmt.Println("proc end")
				}
				break
			}

			if errCount >= maxErrCount {
				fmt.Println("query error found stop")
				break
			}
		}
		fmt.Println("killing proc")
		cronosdCmd.Process.Kill()
		fmt.Println("killed proc")
		cronosdCmd.Process.Wait()
		fmt.Println("proc end")
		file.Close()
		fmt.Println("log closed")

		log, err := os.ReadFile(l)
		if err != nil {
			fmt.Println("read log err:", err)
		} else {
			fmt.Println("checking log")
			if strings.Contains(string(log), errMsg) {
				fmt.Println("mm-found")
				break
			}
		}
	}
}
