package execute

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"os"
	"stock-simulator-serverless/src/logic"
	"stock-simulator-serverless/src/seed"
	"stock-simulator-serverless/src/storage"
	"strings"
	"time"
)

type CommandToolExecutor struct {
	sess       *session.Session
	ddbClient  *dynamodb.DynamoDB
	enviorment string
	tableName  string

	gqlLogGroupName string

	l *logic.Logic
	s *storage.DdbTable
}

func invalidCommand(msg string) error {
	return fmt.Errorf("invalid command err=%v", msg)
}

func New() (*CommandToolExecutor, error) {
	env := flag.String("env", "staging", "the target stack")
	flag.Parse()

	env = aws.String("starket-" + *env)
	cte := &CommandToolExecutor{
		sess:       session.Must(session.NewSession()),
		enviorment: *env,
	}

	cfn := cloudformation.New(cte.sess)
	output, err := cfn.DescribeStacks(&cloudformation.DescribeStacksInput{})
	if err != nil {
		return nil, err
	}
	cte.ddbClient = dynamodb.New(cte.sess)
	var targetStack *cloudformation.Stack
	for _, s := range output.Stacks {
		if *s.StackName == *env {
			targetStack = s
			break
		}
	}

	if targetStack == nil {
		return nil, fmt.Errorf("failed to find stack " + *env)
	}

	for _, o := range targetStack.Outputs {
		parsed := strings.Replace(*o.ExportName, *env+"-", "", 1)
		switch parsed {
		case "item-table":
			cte.tableName = *o.OutputValue
		case "gql-logs":
			cte.gqlLogGroupName = *o.OutputValue
		}
	}

	cte.s = storage.New(cte.tableName, cte.ddbClient)
	cte.l = logic.New(cte.s)

	return cte, nil
}

func (cte *CommandToolExecutor) Exec(args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, invalidCommand("missing args")
	}
	switch args[0] {
	case "storage":
		return cte.storage(args[1:])
	case "logs":
		return cte.logs(args[1:])
	default:
		return nil, invalidCommand(fmt.Sprintf("%v is not a valid command", args[0]))
	}
}

func (cte *CommandToolExecutor) logs(args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, invalidCommand("missing arg")
	}

	startTime := "-5m"
	endTime := ""
	filter := ""
	if len(args) > 2 {
		for i, ele := range args[1:] {
			switch ele {
			case "--start":
				startTime = args[i+2]
			case "--end":
				endTime = args[i+2]
			case "--filter":
				endTime = args[i+2]
			}
		}
	}

	switch args[0] {
	case "gql-lambda":
		err := FollowLogs(cte.gqlLogGroupName, startTime, endTime, filter)
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	return nil, nil
}

func (cte *CommandToolExecutor) warn(message string) {
	_, _ = fmt.Fprint(os.Stderr, "WARNING:", message, "\n")
	_, _ = fmt.Fprint(os.Stderr, "You are making these changes on ", cte.enviorment, "\n")
	_, _ = fmt.Fprint(os.Stderr, "You have 5 seconds to cancel")
	<-time.After(time.Second * 5)
}

func (cte *CommandToolExecutor) storage(args []string) (interface{}, error) {
	if len(args) == 0 {
		return nil, invalidCommand("no option")
	}

	switch args[0] {
	case "reset":
		// reset the ddb to
		cte.warn("this is destructive and will reset the database")
		return nil, cte.resetTable()
	case "seed":
		cte.warn("this is destructive and will reset the database")
		err := cte.resetTable()
		if err != nil {
			return nil, err
		}
		seed.Two(cte.l, cte.s)
		return nil, nil
	default:
		return nil, invalidCommand(fmt.Sprintf("%v is not a valid command", args[0]))
	}
}

func (cte *CommandToolExecutor) resetTable() error {
	_, err := cte.ddbClient.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: &cte.tableName,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}

	for j := 0; j < 100; j++ {
		_, err = cte.ddbClient.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: &cte.tableName,
		})
		<-time.After(time.Millisecond * 200)
		if err != nil {
			break
		}
	}
	table := *storage.StarketTable
	table.TableName = &cte.tableName
	_, err = cte.ddbClient.CreateTable(&table)
	if err != nil {
		return err
	}

	for j := 0; j < 100; j++ {
		tableOutput, err := cte.ddbClient.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: &cte.tableName,
		})
		if err != nil {
			return nil
		}
		if *tableOutput.Table.TableStatus == dynamodb.TableStatusActive {
			break
		}
		<-time.After(time.Millisecond * 500)
	}

	return nil
}
