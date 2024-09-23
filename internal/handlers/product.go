package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"simple-product-elasticsearch/internal/config"
	"simple-product-elasticsearch/internal/models"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v3"
)

// GetProducts 	handles GET requests to fetch products
// @Summary		Get Products
// @Description	Retrieves a list of products with optional pagination and search keywords
// @Tags			Products
// @Accept			json
// @Produce		json
// @Param			limit	query	int		false	"Limit number of results"
// @Param			offset	query	int		false	"Offset for pagination"
// @Param			search	query	string	false	"Search keyword"
// @Success		200		{array}	models.Product
// @Router			/products [get]
func GetProducts(cfg *config.Config, es *elasticsearch.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		limit, offset, search := c.Query("limit", "10"), c.Query("offset", "0"), c.Query("search")
		index := "products"
		m := []models.Product{}

		var r map[string]interface{}
		log.Printf("Search: %v, Offset: %v, Limit: %v", search, offset, limit)

		query := map[string]interface{}{
			"size": limit,
		}

		if search != "" && len(search) > 0 {
			query = map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"should": []map[string]interface{}{
							{
								"bool": map[string]interface{}{
									"should": []map[string]interface{}{
										{
											"multi_match": map[string]interface{}{
												"query":     search,
												"fields":    []string{"product_name", "drug_generic", "company"},
												"operator":  "and",
												"fuzziness": "AUTO",
											},
										},
									},
								},
							},
							{
								"bool": map[string]interface{}{
									"should": []map[string]interface{}{
										{
											"wildcard": map[string]interface{}{
												"product_name": map[string]interface{}{
													"value":            "*" + search + "*",
													"case_insensitive": true,
												},
											},
										},
										{
											"wildcard": map[string]interface{}{
												"drug_generic": map[string]interface{}{
													"value":            "*" + search + "*",
													"case_insensitive": true,
												},
											},
										},
										{
											"wildcard": map[string]interface{}{
												"company": map[string]interface{}{
													"value":            "*" + search + "*",
													"case_insensitive": true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"sort": []map[string]interface{}{
					{"_score": map[string]interface{}{"order": "desc"}},              // sort by _score in descending order
					{"product_name.keyword": map[string]interface{}{"order": "asc"}}, // sort by product_name in ascending order
				},
				"size": limit, // Number of results to return
			}
		}

		queryStr, _ := json.Marshal(query)
		log.Printf("ESQuery=\n: %v", string(queryStr))

		// Build the request body.
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Printf("Error encoding ESQuery: %s", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		// Change the offset to int
		numberOffset, err := strconv.Atoi(offset)
		if err != nil {
			fmt.Println("Error converting string to int:", err)
		}

		// Perform the search request.
		res, err := es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex(index),
			es.Search.WithBody(&buf),
			es.Search.WithTrackTotalHits(true),
			es.Search.WithPretty(),
			es.Search.WithFrom(numberOffset),
		)

		if err != nil {
			log.Printf("Error getting response: %s", err)

		}
		defer res.Body.Close()

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Printf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				errors := fmt.Sprintf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
				log.Print(errors)
				return fiber.NewError(fiber.StatusInternalServerError, errors)
			}
		}

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		if hits, ok := r["hits"].((map[string]interface{}))["hits"]; ok {
			if _, ok := hits.([]interface{}); ok {
				for _, hit := range hits.([]interface{}) {
					docId := hit.(map[string]interface{})["_id"]
					score := hit.(map[string]interface{})["_score"]
					itemSuggestion := hit.(map[string]interface{})["_source"]

					jsonStr, err := json.Marshal(itemSuggestion)
					if err != nil {
						fmt.Println(err)
					}

					var item models.Product
					err = json.Unmarshal(jsonStr, &item)
					if err != nil {
						fmt.Println(err)
					}

					id, _ := strconv.Atoi(docId.(string))
					item.ID = uint64(id)
					item.Score = score.(float64)
					m = append(m, item)
				}
			}
		}

		// Print the response status, number of results, and request duration.
		log.Printf(
			"[%s] %d hits; took: %dms",
			res.Status(),
			int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(r["took"].(float64)),
		)

		c.Response().Header.SetContentType("application/json")
		return c.JSON(m)
	}
}
