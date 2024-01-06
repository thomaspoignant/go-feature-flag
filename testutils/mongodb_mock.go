package testutils

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var mongoFindResultString = `[{"flag":"test-flag","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},{"flag":"test-flag2","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var mongoFindResultNoFlag = `[{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var mongoFindResultFlagNoStr = `[{"flag":123456,"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var QueryResult = `{"test-flag":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},"test-flag2":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}}`

func mapToBsonD(inputMap []map[string]interface{}) []bson.D {
	bsonData := make([]bson.D, 0, 10)

	for _, v := range inputMap {
		var b []bson.E

		for k, v := range v {
			b = append(b, bson.E{Key: k, Value: v})
		}

		bsonData = append(bsonData, bson.D(b))
	}

	return bsonData
}

func succesQueryMockerFactory(queryResult string) func(t *mtest.T) {
	return func(t *mtest.T) {
		var resultMap []map[string]interface{}

		// Unmarshal the JSON string into the map
		err := json.Unmarshal([]byte(queryResult), &resultMap)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Convert the map to a BSON document (bson.D)
		bsonDocuments := mapToBsonD(resultMap)

		t.AddMockResponses(mtest.CreateCursorResponse(1, "dummyDB.dummyCollection",
			mtest.FirstBatch,
			bsonDocuments...,
		))
	}
}

var MockSuccessFind = succesQueryMockerFactory(mongoFindResultString)

var MockNoFlagKeyResult = succesQueryMockerFactory(mongoFindResultNoFlag)

var MockFlagNotStrResult = succesQueryMockerFactory(mongoFindResultFlagNoStr)

var MockNoFlags = func(t *mtest.T) {

	t.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{}))
}
