#!/bin/bash

cust_func(){
  for i in {1..1000}
  do
    echo "$i"
      curl -X GET "https://jazz-adapter-api-test.k8s.bestcorp.net/style/12200509" -H  "accept: text/plain" &
  done
}

for x in {1..5}
do
  echo "loop number $x"
	cust_func & # Put a function in the background
done

wait 
printf "\nAll done\n"



# echo "POST http://localhost:8080/\nContent-Type: application/json\n@/Users/lasse.gaardsholt/load-test.json" | vegeta attack -duration=10s -output=create.bin
# vegeta plot -title="Create Results" create.bin > create.html
# open create.html

