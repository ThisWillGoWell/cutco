package storage

import (
	local_ddb "stock-simulator-serverless/local-ddb"
	"testing"
)

func NewLocalDdb() *DdbTable {
	i := local_ddb.Instance()
	err := i.Cleanup(StarketTable)
	if err != nil {
		panic(err)
	}
	return New(*StarketTable.TableName, i.DdbClient)
}

func NewTestingDdb(t *testing.T) *DdbTable {
	i := local_ddb.Instance()
	err := i.Cleanup(StarketTable)
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		i.Shutdown()
	})
	return New(*StarketTable.TableName, i.DdbClient)
}
