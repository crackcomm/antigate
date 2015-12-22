package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/crackcomm/antigate"
)

var (
	captchaURL  = flag.String("captcha-url", "", "Captcha URL")
	antigateKey = flag.String("antigate-key", "", "Antigate API Key (also accepts ANTIGATE_KEY environment variable)")
	outputImage = flag.String("output-image", "", "Output image (optional)")
)

func init() {
	log.SetFlags(0)
}

func main() {
	flag.Parse()

	if *antigateKey == "" {
		if env := os.Getenv("ANTIGATE_KEY"); env != "" {
			*antigateKey = env
		} else {
			log.Fatal("Antigate key is expected in --antigate-key flag or ANTIGATE_KEY environment variable")
		}
	}

	client := &antigate.Client{
		Key:           *antigateKey,
		MaxRetries:    35,
		CheckInterval: 500 * time.Millisecond,
		RetryInterval: 500 * time.Millisecond,
	}

	resp, err := http.Get(*captchaURL)
	if err != nil {
		log.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	image, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("read error: %v", err)
	}

	if *outputImage != "" {
		err = ioutil.WriteFile(*outputImage, image, os.ModePerm)
		if err != nil {
			log.Fatalf("save error: %v", err)
		}
	}

	text, err := client.Solve(image)
	if err != nil {
		log.Fatalf("solve error: %v", err)
	}

	log.Println(text)
}
