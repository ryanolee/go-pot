package main

import (
	//	"github.com/ryanolee/ryan-pot/cmd"
	"encoding/json"
	"fmt"

	"github.com/ryanolee/ryan-pot/generator/chaff"
	"github.com/ryanolee/ryan-pot/rand"
	//"github.com/ryanolee/ryan-pot/rand"
)



func main() {
  res, _ := chaff.ParseSchemaFile("test-schema.json")
  //fmt.Println(res, err)
  //fmt.Println(res.Metadata.Errors)

  // Parellelize this
  data, _ := json.Marshal(res.Generate(&chaff.GeneratorOptions{
    Rand: rand.NewSeededRandFromTime(),
    MaximumReferenceDepth: 100,
    BypassCyclicReferenceCheck: true,
  }))

  fmt.Println(string(data))
    
  
  //

  
  //cmd.Execute()
}

//func main() {
	//rules := secrets.GetGenerators()
	//for _, rule := range rules {
	//	fmt.Printf("%s: %s = %s\n", rule.Name, rule.NameGenerator.Generate(), rule.SecretGenerator.Generate())
	//}
//}