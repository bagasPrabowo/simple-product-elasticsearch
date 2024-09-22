# Simple Product Elasticsearch
This project is skill test for one of Indonesian Company.

## How To Run
1. Create .env file Fill the value with key in .env.example or Copy the .env.example with
    > cp .env.example .env

2. Simply run
    > docker compose up -d

3. If you want migrate sample data, you can use template
    ```
    {"index":{"_index": "{index_name}", "_id": {(optional)}}}
    {...} // json data to be restore
    ```
    or you can use the products.json file that i already prepare and run
    > curl -X POST "http://localhost:9200/products/_bulk" -H 'Content-Type: application/json' --data-binary @internal/seeder/products.json

4. Run the project
    > go run cmd/main.go

5. If you want to run locally (without docker), please make sure the env value is correct

## Swagger Docs
http://localhost:8080/swagger/index.html
