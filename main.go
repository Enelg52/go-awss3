package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	bucket := flag.String("b", "sto1.team.00", "Bucket name")
	file := flag.String("f", "test2000", "File name")
	mode := flag.String("m", "list", "Mode: list, upload, download")
	flag.Parse()
	// Initialize a session in eu-south-1 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("eu-south-1")},
	)
	// Create S3 service client
	svc := s3.New(sess)
	// Create s3 uploader
	uploader := s3manager.NewUploader(sess)
	// Create s3 downloader
	downloader := s3manager.NewDownloader(sess)

	switch *mode {
	case "list":
		err := printObjects(*svc, *bucket)
		if err != nil {
			exitError(err)
		}
	case "upload":
		hash, err := getMD5(*file)
		if err != nil {
			exitError(err)
		}
		err = upload(uploader, *file, *bucket)
		if err != nil {
			exitError(err)
		}
		hashS3, err := getS3MD5(*svc, *file, *bucket)
		if err != nil {
			exitError(err)
		}
		if hash != hashS3 {
			fmt.Println("[*] Error, the checksum is wrong")
			fmt.Println(hash)
			fmt.Println(hashS3)
		} else {
			fmt.Println("[*] Checksum is right !")
		}
	case "download":
		hashS3, err := getS3MD5(*svc, *file, *bucket)
		if err != nil {
			exitError(err)
		}
		err = download(downloader, *file, *bucket)
		if err != nil {
			exitError(err)
		}
		hash, err := getMD5(*file)
		if err != nil {
			exitError(err)
		}
		if hash != hashS3 {
			fmt.Println("[*] Error, the checksum is wrong")
			fmt.Println(hash)
			fmt.Println(hashS3)
		} else {
			fmt.Println("[*] Checksum is right !")
		}
	}
}

func exitError(err error) {
	log.Fatalln("[!] Error:", err)
	os.Exit(1)
}

/*func printBuckets(svc s3.S3) error {
	result, err := svc.ListBuckets(nil)
	if err != nil {
		return err
	}
	for _, b := range result.Buckets {
		fmt.Println(aws.StringValue(b.Name))
	}
	return error(nil)
}*/

func upload(uploader *s3manager.Uploader, filename string, bucket string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   f, //file
	})
	if err != nil {
		return err
	}
	fmt.Println("[*] Upload worked !")
	return error(nil)
}

func download(downloader *s3manager.Downloader, item string, bucket string) error {
	file, err := os.Create(item)
	if err != nil {
		exitError(err)
	}
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		return err
	}
	fmt.Println("[*] Downloaded", file.Name(), numBytes, "bytes")
	return error(nil)
}

func getS3MD5(svc s3.S3, obj string, bucket string) (string, error) {
	l, err := svc.HeadObject(&s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(obj)})
	if err != nil {
		return "", err
	}
	hash := *l.ETag
	return strings.Trim(hash, "\""), nil
}

func printObjects(svc s3.S3, bucket string) error {
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil {
		return err
	}
	fmt.Println("[*] Objects in bucket", bucket, "\n")
	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("")
	}
	return error(nil)
}

func getMD5(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil

}
