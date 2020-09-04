package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/hashicorp/logutils"
	vault "github.com/mch1307/vaultlib"
)

type vaultClient interface {
	GetSecret(string) (vault.Secret, error)
	IsAuthenticated() bool
}

var	vaultCli vaultClient = nil
var vaultCache map[string]string = map[string]string{}

func main() {
	var args struct {
		InPlace	bool     `arg:"-i,--inplace" help:"Overwrite input files"`
		Quiet   bool     `arg:"-q,--quiet" help:"Suppress output (useful when writing files in place)"`
		Debug   bool		 `arg:"-d,--debug" help:"Enable debug logging" default:"false"`
		Files   []string `arg:"positional,required" help:"List of files to run against"`
	}
	arg.MustParse(&args)

	initializeLogger(args.Debug)
	err := initializeVault()

	if err != nil {
		log.Fatal("[ERROR] Failed to initialize Vault Client ", err)
	}

	log.Println("[DEBUG] Interpolating the following files: ", args.Files)

	for _, filePath := range args.Files {
	  fileHandle, err := os.Open(filePath)

		if err != nil {
		  log.Fatal("[ERROR] Failed to open file ", filePath, err)
		}

		fileScanner := bufio.NewScanner(fileHandle)

		output := []string{}
		for fileScanner.Scan() {
			line := fileScanner.Text()

			substitutions := parseTokens(line)

			if substitutions != nil {
				log.Println("[DEBUG] Found token in line: " + line)
				newLine, err := performSubstitutions(line, substitutions)

				if err != nil {
				  _ = fileHandle.Close()
				  log.Fatal("[ERROR] Failed to interpolate values", err)
				}
				log.Println("[DEBUG]Interpolated line:", newLine)

				output = append(output, newLine)
			} else {
				output = append(output, line)
			}
		}

		err = fileHandle.Close()

		if err != nil {
			log.Println("[ERROR] Failed to close file " + filePath, err)
		}

		var fileWriter *bufio.Writer
		if args.InPlace {
			fileHandle, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
			fileWriter = bufio.NewWriter(fileHandle)
		}

		for _, line := range output {
			if !args.Quiet {
				fmt.Println(line)
			}
			if args.InPlace {
				_, _ = fileWriter.WriteString(line + "\n")
			  fileWriter.Flush()
			}
		} 

		if args.InPlace {
		  err = fileHandle.Close()

		  if err != nil {
				log.Fatal("[ERROR] Failed to close file " + filePath, err)
		  }
		}
	}
}	

func initializeLogger(debug bool) {
	logFilter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer: os.Stderr,
	}

	if debug {
		logFilter.MinLevel = logutils.LogLevel("DEBUG")
	}

	log.SetOutput(logFilter)
}

func initializeVault() error {
	var err error
	vcConf := vault.NewConfig()

	vaultCli, err = vault.NewClient(vcConf)

	if err != nil {
		return err
	}

	if ! vaultCli.IsAuthenticated() {
		return errors.New("Couldn't communicate with Vault: Unauthorized")
	}

	return nil
}

func parseTokens(line string) []string {
	return regexp.MustCompile(`\(\( .*/.*\)\)`).FindAllString(line, -1)
}

func performSubstitutions(line string, substitutions []string) (string, error) {
	newLine := line
	for _, s := range substitutions {
		log.Println("[DEBUG] Found value to interpolate:", s)

		if _, exists := vaultCache[s]; !exists {
	    secretDef := stripParens(s)
			log.Println("[DEBUG] Interpolating secret: ", secretDef)
	    secretPath, secretKey := parseSecret(secretDef)
			log.Println("[DEBUG] Secret Path: ", secretPath)
			log.Println("[DEBUG] Secret Key: ", secretKey)

			secret, err := vaultCli.GetSecret(secretPath)
			log.Println("[DEBUG] Retrieved Secret: ", secret.KV)

			if err != nil {
				return "", err
			}

			vaultCache[s] = secret.KV[secretKey]
			log.Println("[DEBUG] ", vaultCache)
		} else {
			log.Println("[DEBUG] Secret ", s, " already in cache")
		}

	  newLine = strings.ReplaceAll(newLine, s, vaultCache[s])
	}

	return newLine, nil
}

func stripParens(s string) string {
	return string(regexp.MustCompile(`\(\( *| *\)\)`).ReplaceAll([]byte(s), []byte("")))
}

func parseSecret(secret string) (string, string) {
	var secretKey string = "value"

	secretPath := regexp.MustCompile(`:`).Split(secret, -1)

	if len(secretPath) == 2 {
		secretKey = secretPath[1]
	}	

	return secretPath[0], secretKey
}