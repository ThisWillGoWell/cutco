package local_ddb

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var instance *LocalDDB

type LocalDDB struct {
	process   *exec.Cmd
	DdbClient *dynamodb.DynamoDB
}

func Instance() *LocalDDB {
	if instance == nil {
		var err error
		instance, err = newInLocalDDB()
		if err != nil {
			panic(err)
		}
	}
	return instance
}

func start(port int) (*exec.Cmd, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var seperator string
	switch runtime.GOOS {
	case "darwin":
		seperator = `/`
	case "windows":
		seperator = `\`
	}

	pathList := strings.Split(dir, seperator)
	for i, s := range pathList {
		if s == "stock-simulator-serverless" {
			pathList = pathList[0 : i+1]
			break
		}
	}
	path := strings.Join(pathList, seperator) + seperator + `local-ddb` + seperator
	libPath := "-Djava.library.path=" + path + "DynamoDBLocal_lib"
	jarPath := path + "DynamoDBLocal.jar"

	process := exec.Command("java", libPath, "-jar", jarPath, "-port", fmt.Sprintf("%d", port), "-inMemory", "-sharedDb")
	//process.Stdout = os.Stdout
	//process.Stderr = os.Stderr

	err = process.Start()
	if err != nil {
		return nil, err
	}

	// termination
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-sigs

	}()
	return process, nil
}

func (local *LocalDDB) Shutdown() {
	if local.process == nil {
		return
	}
	err := local.process.Process.Kill()
	if err != nil {
		fmt.Printf("failed to close ddb err=%v", err)
	}
	for i := 0; i < 4; i++ {
		state, err := local.process.Process.Wait()
		if err != nil || state.Exited() {
			fmt.Println("killed ddb")
			return
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func running() bool {
	// default port 9000
	l1, err := net.Listen("tcp", ":9000")
	if err == nil {
		_ = l1.Close()
		return false
	}
	return true

}

func (local *LocalDDB) Cleanup(input *dynamodb.CreateTableInput) error {
	result, err := local.DdbClient.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		return err
	}
	for _, t := range result.TableNames {
		_, err := local.DdbClient.DeleteTable(&dynamodb.DeleteTableInput{
			TableName: t,
		})
		if err != nil {
			return err
		}
	}
	// make a new table with the new schema if provided
	if input != nil {
		_, err = local.DdbClient.CreateTable(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func newInLocalDDB() (*LocalDDB, error) {
	var process *exec.Cmd
	if !running() {
		var err error
		process, err = start(9000)
		if err != nil {
			return nil, err
		}
	}

	localDdb := &LocalDDB{
		process: process,
	}

	address := "http://localhost:9000"

	cfg := &aws.Config{
		Endpoint:    aws.String(address),
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials("key", "secret;", ""),
	}
	s, _ := session.NewSession()
	ddb := dynamodb.New(s, cfg)
	ddb.Endpoint = address

	success := false
	// wait for table to come up
	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		_, err := ddb.ListTablesWithContext(ctx, &dynamodb.ListTablesInput{})
		cancel()
		if err == nil {
			success = true
			break
		}
	}
	if !success {
		return nil, fmt.Errorf("failed to connect")
	}
	localDdb.DdbClient = ddb
	return localDdb, nil
}
