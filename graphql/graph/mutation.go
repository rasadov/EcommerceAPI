package graph

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/rasadov/EcommerceAPI/graphql/generated"
	"github.com/rasadov/EcommerceAPI/order/models"
	payment "github.com/rasadov/EcommerceAPI/payment/proto/pb"
	"github.com/rasadov/EcommerceAPI/pkg/auth"
	"github.com/rasadov/EcommerceAPI/pkg/middleware"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
)

type mutationResolver struct {
	server *Server
}

func (resolver *mutationResolver) Register(ctx context.Context, in generated.RegisterInput) (*generated.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	token, err := resolver.server.accountClient.Register(ctx, in.Name, in.Email, in.Password)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ginContext, ok := middleware.GinContextFromContext(ctx)
	if !ok {
		return nil, errors.New("could not retrieve gin context")
	}
	ginContext.SetCookie("token", token, 3600, "/", "localhost", false, true)

	return &generated.AuthResponse{Token: token}, nil
}

func (resolver *mutationResolver) Login(ctx context.Context, in generated.LoginInput) (*generated.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	token, err := resolver.server.accountClient.Login(ctx, in.Email, in.Password)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ginContext, ok := middleware.GinContextFromContext(ctx)
	if !ok {
		return nil, errors.New("could not retrieve gin context")
	}
	ginContext.SetCookie("token", token, 3600, "/", "localhost", false, true)

	return &generated.AuthResponse{Token: token}, nil
}

func (resolver *mutationResolver) CreateProduct(ctx context.Context, in generated.CreateProductInput) (*generated.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	log.Println("CreateProduct called with input:", in)

	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		return nil, err
	}
	log.Println("CreateProduct called with accountId:", accountId)
	postProduct, err := resolver.server.productClient.PostProduct(ctx, in.Name, in.Description, in.Price, int64(accountId))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("Created product:", postProduct)
	log.Println("Product id: ", postProduct.ID)

	return &generated.Product{
		ID:          postProduct.ID,
		Name:        postProduct.Name,
		Description: postProduct.Description,
		Price:       postProduct.Price,
		AccountID:   accountId,
	}, nil
}

func (resolver *mutationResolver) UpdateProduct(ctx context.Context, in generated.UpdateProductInput) (*generated.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		return nil, err
	}

	updatedProduct, err := resolver.server.productClient.UpdateProduct(ctx, in.ID, in.Name, in.Description, in.Price, int64(accountId))
	if err != nil {
		return nil, err
	}
	return &generated.Product{
		ID:          updatedProduct.ID,
		Name:        updatedProduct.Name,
		Description: updatedProduct.Description,
		Price:       updatedProduct.Price,
		AccountID:   accountId,
	}, nil
}

func (resolver *mutationResolver) DeleteProduct(ctx context.Context, id string) (*bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		return nil, err
	}

	err = resolver.server.productClient.DeleteProduct(ctx, id, int64(accountId))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	success := true
	return &success, nil
}

func (resolver *mutationResolver) CreateOrder(ctx context.Context, in generated.OrderInput) (*generated.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var products []*models.OrderedProduct
	for _, product := range in.Products {
		if product.Quantity <= 0 {
			return nil, ErrInvalidParameter
		}
		products = append(products, &models.OrderedProduct{
			ID:       product.ID,
			Quantity: uint32(product.Quantity),
		})
	}

	accountId, err := auth.GetUserIdInt(ctx, true)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	postOrder, err := resolver.server.orderClient.PostOrder(ctx, uint64(accountId), products)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var orderedProducts []*generated.OrderedProduct
	for _, orderedProduct := range postOrder.Products {
		orderedProducts = append(orderedProducts, &generated.OrderedProduct{
			ID:          orderedProduct.ID,
			Name:        orderedProduct.Name,
			Description: orderedProduct.Description,
			Price:       orderedProduct.Price,
			Quantity:    int(orderedProduct.Quantity),
		})
	}

	return &generated.Order{
		ID:         int(postOrder.ID),
		CreatedAt:  postOrder.CreatedAt,
		TotalPrice: postOrder.TotalPrice,
		Products:   orderedProducts,
	}, nil
}

func (resolver *mutationResolver) CreateCustomerPortalSession(ctx context.Context, credentials *generated.CustomerPortalSessionInput) (*generated.RedirectResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	UrlWithSession, err := resolver.server.paymentClient.CreateCustomerPortalSession(ctx, uint64(credentials.AccountID), credentials.Email, credentials.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &generated.RedirectResponse{URL: UrlWithSession}, nil
}

func (resolver *mutationResolver) CreateCheckoutSession(ctx context.Context, details *generated.CheckoutInput) (*generated.RedirectResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var products []*payment.CartItem
	for _, product := range details.Products {
		products = append(products, &payment.CartItem{
			ProductId: product.ID,
			Quantity:  uint64(product.Quantity),
		})
	}

	UrlWithCheckoutSession, err := resolver.server.paymentClient.CreateCheckoutSession(ctx, details.OrderID,
		details.AccountID, details.Email, details.Name, details.RedirectURL, products)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &generated.RedirectResponse{URL: UrlWithCheckoutSession}, nil
}
