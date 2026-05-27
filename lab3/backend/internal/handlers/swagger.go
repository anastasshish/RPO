package handlers

import "github.com/gin-gonic/gin"

func SwaggerPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `<!doctype html><html><head><title>Transport Auth API Swagger</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/api/v1/openapi.json',dom_id:'#swagger-ui'});</script></body></html>`)
}
func OpenAPI(c *gin.Context) {
	c.JSON(200, gin.H{"openapi": "3.0.0", "info": gin.H{"title": "Transport Card Payment Authorization API", "version": "1.0.0"}, "servers": []gin.H{{"url": "/api/v1"}}, "paths": gin.H{"/auth/login": gin.H{"post": gin.H{"summary": "Login by password and receive JWT"}}, "/terminal/payments/authorize": gin.H{"post": gin.H{"summary": "Authorize terminal payment transaction"}}, "/terminal/keys": gin.H{"get": gin.H{"summary": "Download all MIFARE keys for terminal"}}, "/cards": gin.H{"get": gin.H{"summary": "List cards"}, "post": gin.H{"summary": "Create card"}}, "/terminals": gin.H{"get": gin.H{"summary": "List terminals"}, "post": gin.H{"summary": "Create terminal"}}, "/transactions": gin.H{"get": gin.H{"summary": "List transactions"}}, "/keys": gin.H{"get": gin.H{"summary": "List keys admin only"}}, "/users": gin.H{"get": gin.H{"summary": "List users admin only"}}}})
}
