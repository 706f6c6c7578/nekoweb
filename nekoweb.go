package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/net/proxy"
)

func main() {
	download := flag.Bool("d", false, "Download file from URL")
	useTor := flag.Bool("T", false, "Use Tor network")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: nekoweb [-d] [-T] [folder/]filename or URL")
		os.Exit(1)
	}

	var client *http.Client
	if *useTor {
		dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
		if err != nil {
			fmt.Println("Error creating SOCKS5 dialer:", err)
			os.Exit(1)
		}
		httpTransport := &http.Transport{Dial: dialer.Dial}
		client = &http.Client{Transport: httpTransport}
	} else {
		client = &http.Client{}
	}

	if *download {
		url := flag.Args()[0]
		err := downloadFile(url, client)
		if err != nil {
			fmt.Println("Error downloading file:", err)
			os.Exit(1)
		}
	} else {
		filepathArg := flag.Args()[0]
		err := uploadFile(filepathArg, client)
		if err != nil {
			fmt.Println("Error uploading file:", err)
			os.Exit(1)
		}
	}
}

func downloadFile(url string, client *http.Client) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func uploadFile(filepathArg string, client *http.Client) error {
	folder := filepath.Dir(filepathArg)
	filename := filepath.Base(filepathArg)

	if folder == "." {
		folder = "/"
	}

	fileContent, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return fmt.Errorf("error creating form file: %w", err)
	}

	_, err = io.Copy(part, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("error copying file content: %w", err)
	}

	err = writer.WriteField("pathname", folder)
	if err != nil {
		return fmt.Errorf("error writing field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error closing writer: %w", err)
	}

	req, err := http.NewRequest("POST", "https://nekoweb.org/api/files/upload", body)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", "your-nekoweb-api-key")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	fmt.Println("Response:", string(respBody))
	return nil
}
