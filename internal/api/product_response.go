package api

import "user-frontend/internal/model"

type productResponse struct {
	model.Product
	ImageURL string `json:"image_url"`
}

func buildProductResponse(product *model.Product) productResponse {
	if product == nil {
		return productResponse{}
	}
	response := productResponse{Product: *product}
	if ProductImageSvc == nil {
		return response
	}
	image, err := ProductImageSvc.GetPrimaryImage(product.ID)
	if err == nil && image != nil {
		response.ImageURL = image.URL
	}
	return response
}

func buildProductResponses(products []model.Product) []productResponse {
	responses := make([]productResponse, 0, len(products))
	for i := range products {
		responses = append(responses, buildProductResponse(&products[i]))
	}
	return responses
}
