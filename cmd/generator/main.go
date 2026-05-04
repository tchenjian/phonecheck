package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/bits-and-blooms/bloom/v3"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	inputFile := flag.String("i", "phone_numbers.txt", "input txt file with phone numbers (one per line)")
	outputType := flag.String("t", "all", "output type: bitmap, bloom, sqlite, all")
	outputPrefix := flag.String("o", "phone_numbers", "output file prefix")
	flag.Parse()

	if *inputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Reading phone numbers from %s...\n", *inputFile)
	phones, err := readPhoneNumbers(*inputFile)
	if err != nil {
		fmt.Printf("Failed to read phone numbers: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded %d phone numbers\n", len(phones))

	types := []string{"bitmap", "bloom", "sqlite"}
	if *outputType != "all" {
		types = strings.Split(*outputType, ",")
	}

	for _, t := range types {
		switch t {
		case "bitmap":
			err = generateBitmap(phones, *outputPrefix+".bin")
		case "bloom":
			err = generateBloom(phones, *outputPrefix+"_bloom.bin")
		case "sqlite":
			err = generateSqlite(phones, *outputPrefix+".db")
		default:
			fmt.Printf("Unknown type: %s, skipped\n", t)
			continue
		}
		if err != nil {
			fmt.Printf("Failed to generate %s: %v\n", t, err)
		}
	}

	fmt.Println("Done!")
}

func readPhoneNumbers(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var phones []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			phones = append(phones, line)
		}
	}
	return phones, scanner.Err()
}

func generateBitmap(phones []string, outputFile string) error {
	fmt.Printf("Generating Roaring Bitmap to %s...\n", outputFile)
	bitmap := roaring64.NewBitmap()
	for _, phone := range phones {
		phoneNum, err := strconv.ParseUint(phone, 10, 64)
		if err != nil {
			fmt.Printf("Skipping invalid phone number: %s\n", phone)
			continue
		}
		bitmap.Add(phoneNum)
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = bitmap.WriteTo(outFile)
	if err != nil {
		return err
	}
	fmt.Printf("  Bitmap cardinality: %d\n", bitmap.GetCardinality())
	return nil
}

func generateBloom(phones []string, outputFile string) error {
	fmt.Printf("Generating Bloom Filter to %s...\n", outputFile)
	n := uint(len(phones))
	fpRate := 0.01
	filter := bloom.NewWithEstimates(n, fpRate)

	for _, phone := range phones {
		filter.AddString(phone)
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = filter.WriteTo(outFile)
	if err != nil {
		return err
	}
	fmt.Printf("  Bloom filter: %d items, fp rate: %.2f%%\n", n, fpRate*100)
	return nil
}

func generateSqlite(phones []string, outputFile string) error {
	fmt.Printf("Generating SQLite DB to %s...\n", outputFile)

	os.Remove(outputFile)
	db, err := sql.Open("sqlite3", outputFile)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE phones (number TEXT PRIMARY KEY)`)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO phones (number) VALUES (?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, phone := range phones {
		_, err = stmt.Exec(phone)
		if err != nil {
			fmt.Printf("Skipping duplicate phone number: %s\n", phone)
			continue
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Printf("  SQLite DB: %d records\n", len(phones))
	return nil
}
