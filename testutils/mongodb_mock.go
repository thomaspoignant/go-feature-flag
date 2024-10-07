package testutils

var MongoFindResultString = `[{"flag":"test-flag","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},{"flag":"test-flag2","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var MongoMissingFlagKey = `[{"flag":"test-flag","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var MissingFlagKeyResult = `{"test-flag":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}}`

var MongoFindResultFlagNoStr = `[{"flag":123456,"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},{"flag":"test-flag","variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}]`

var FlagKeyNotStringResult = MissingFlagKeyResult

var QueryResult = `{"test-flag":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"random-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false},"test-flag2":{"variations":{"true_var":true,"false_var":false},"targeting":[{"query":"key eq \"not-a-key\"","percentage":{"true_var":0,"false_var":100}}],"defaultRule":{"variation":"false_var"},"trackEvents":false}}`
