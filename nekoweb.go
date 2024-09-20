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
)

func main() {
	download := flag.Bool("d", false, "Download file from URL")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: nekoweb [-d] [folder/]filename or URL")
		os.Exit(1)
	}

	if *download {
		url := flag.Args()[0]
		err := downloadFile(url)
		if err != nil {
			fmt.Println("Error downloading file:", err)
			os.Exit(1)
		}
	} else {
		filepathArg := flag.Args()[0]
		err := uploadFile(filepathArg)
		if err != nil {
			fmt.Println("Error uploading file:", err)
			os.Exit(1)
		}
	}
}

func downloadFile(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filename := filepath.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("File downloaded:", filename)
	return nil
}

func uploadFile(filepathArg string) error {
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

	client := &http.Client{}
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
