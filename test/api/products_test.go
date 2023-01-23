package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/product"
)

func Test_Create_Product_Creats_New_Product_In_Database_With_Valid_Input(t *testing.T) {
	// Arrange
	createProductModel := product.CreateProductModel{
		SKU:         "test-sku",
		Name:        "test-name",
		Description: "test-description",
		Price:       1001,
	}

	requestBody := product.CreateProductCommand{
		Product: createProductModel,
	}

	payload, err := json.Marshal(requestBody)

	// Act
	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/products"),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	if err != nil {
		t.Errorf("unexpected error occurred: %s", err.Error())
	}

	if resp == nil {
		t.Errorf("response is nil")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("unexpected error occurred: %s", err.Error())
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status code: %d received: %d", http.StatusCreated, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		t.Errorf("unexpected value for location header: '%s'", location)
	}

	id := path.Base(location)
	var product product.Product
	if err := fixture.db.Get(&product, "SELECT * FROM product WHERE id = $1;", id); err != nil {
		t.Errorf("unexpected error occurred: %s", err.Error())
	}

	if product.SKU != createProductModel.SKU {
		t.Errorf("unexpected value for product SKU expected '%s' found '%s'", createProductModel.SKU, product.SKU)
	}

	if product.Price != createProductModel.Price {
		t.Errorf("unexpected value for product Price expected '%d' found '%d'", createProductModel.Price, product.Price)
	}

	if product.Name != createProductModel.Name {
		t.Errorf("unexpected value for product Name expected '%s' found '%s'", createProductModel.Name, product.Name)
	}

	if product.Description != createProductModel.Description {
		t.Errorf("unexpected value for product Description expected '%s' found '%s'", createProductModel.Description, product.Description)
	}
}
