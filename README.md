**To run the project:**
* install latest go version for your OS from https://go.dev/dl/
* put your safe browsing API key into config/integrations/safebrowsing.json
* run go run . from project root


**To hit the enrich api**

    run curl -X POST -H "Content-Type: application/json" -d @./sample_data/alerts.json http://localhost:8080/enrich