package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Key to use when setting the gin context.
type ginContextKeyType struct{}

// GinContextKey is the context key used to store the Gin context for GraphQL resolvers.
var GinContextKey = ginContextKeyType{}

func GinContextFromContext(ctx context.Context) (*gin.Context, bool) {
	ginContext, ok := ctx.Value(GinContextKey).(*gin.Context)
	return ginContext, ok
}

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Put the gin.Context into the request context so gqlgen can retrieve it
		ctx := context.WithValue(c.Request.Context(), GinContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
