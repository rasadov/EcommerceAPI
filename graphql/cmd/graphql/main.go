package main

import (
	"log"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/rasadov/EcommerceAPI/graphql/config"
	"github.com/rasadov/EcommerceAPI/graphql/graph"
	"github.com/rasadov/EcommerceAPI/pkg/middleware"
)

func newGraphQLHandler(es graphql.ExecutableSchema, enableIntrospection bool) *handler.Server {
	srv := handler.New(es)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	if enableIntrospection {
		srv.Use(extension.Introspection{})
	}

	return srv
}

func main() {
	server, err := graph.NewGraphQLServer(config.AccountUrl, config.ProductUrl, config.OrderUrl, config.PaymentUrl, config.RecommenderUrl)
	if err != nil {
		log.Fatal(err)
	}

	srv := newGraphQLHandler(server.ToExecutableSchema(), config.EnableIntrospection)
	if config.EnableIntrospection {
		log.Printf("GraphQL introspection enabled (environment: %s)\n", config.Environment)
	} else {
		log.Printf("GraphQL introspection disabled (environment: %s)\n", config.Environment)
	}

	engine := gin.Default()

	engine.Use(middleware.GinContextToContextMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "It works",
		})
	})
	engine.POST("/graphql",
		middleware.AuthorizeJWT(),
		gin.WrapH(srv),
	)
	engine.GET("/graphql",
		middleware.AuthorizeJWT(),
		gin.WrapH(srv),
	)
	engine.OPTIONS("/graphql",
		gin.WrapH(srv),
	)
	engine.GET("/playground", gin.WrapH(playground.Handler("Playground", "/graphql")))

	log.Fatal(engine.Run(":8080"))
}
