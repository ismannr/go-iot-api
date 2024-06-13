package service

import (
	"encoding/json"
	"fmt"
	"gin-crud/response"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func GetProvinceList(c *gin.Context) {
	res, err := http.Get("https://ismannr.github.io/api-wilayah-indonesia/api/provinces.json")
	if err != nil {
		response.GlobalResponse(c, "Error retrieving province list", 404, nil)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		response.GlobalResponse(c, "Error reading response body", 404, nil)
		return
	}

	var provinces []map[string]interface{}
	err = json.Unmarshal(body, &provinces)
	if err != nil {
		response.GlobalResponse(c, "Error parsing JSON", 500, nil)
		return
	}
	response.GlobalResponse(c, "Successfully retrieving province list", 200, provinces)
}

func GetCityDependsOnProvince(c *gin.Context) {
	id := c.Param("id")
	res, err := http.Get(fmt.Sprintf("https://ismannr.github.io/api-wilayah-indonesia/api/regencies/%s.json", id))
	if err != nil {
		response.GlobalResponse(c, "Error retrieving city list", 404, nil)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		response.GlobalResponse(c, "Error reading response body", 404, nil)
		return
	}

	var cities []map[string]interface{}
	err = json.Unmarshal(body, &cities)
	if err != nil {
		response.GlobalResponse(c, "Error parsing JSON", 500, nil)
		return
	}
	response.GlobalResponse(c, "Successfully retrieving city list", 200, cities)
}

func GetCity(c *gin.Context) {
	id := c.Param("id")
	res, err := http.Get(fmt.Sprintf("https://ismannr.github.io/api-wilayah-indonesia/api/regency/%s.json", id))
	if err != nil {
		response.GlobalResponse(c, "Error retrieving city", http.StatusNotFound, nil)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		response.GlobalResponse(c, "Error retrieving city: "+res.Status, res.StatusCode, nil)
		return
	}

	var city map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&city); err != nil {
		response.GlobalResponse(c, "Error parsing JSON", http.StatusInternalServerError, nil)
		return
	}

	response.GlobalResponse(c, "Successfully retrieving city", http.StatusOK, city)
}
