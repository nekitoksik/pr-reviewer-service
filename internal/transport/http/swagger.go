package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func registerSwagger(r *gin.Engine) {

	r.StaticFile("/swagger/openapi.yml", "api/openapi.yml")
	r.GET("/swagger", func(c *gin.Context) {
		html := `
<!DOCTYPE html>
<html>
<head>
  <title>Swagger UI</title>
  <link href="https://unpkg.com/swagger-ui-dist@4/swagger-ui.css" rel="stylesheet"/>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
<script>
  window.onload = function() {
    SwaggerUIBundle({
      url: "/swagger/openapi.yml",
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis],
      layout: "BaseLayout",
    });
  };
</script>
</body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})
}
