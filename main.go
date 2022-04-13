package main

import (
	"github.com/amarchese96/puccini-server/puccini"
	"github.com/gin-gonic/gin"
	"net/http"
)

type compileRequest struct {
	ServiceTemplateUrl string `json:"serviceTemplateUrl"`
	ScriptletName      string `json:"scriptletName,omitempty"`
	ScriptletUrl       string `json:"scriptletUrl,omitempty"`
	Inputs             string `json:"inputs,omitempty"`
}

type compileResponse struct {
	Output       string `json:"output,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func compileFromUrl(c *gin.Context) {
	var compileRequest compileRequest

	if err := c.BindJSON(&compileRequest); err != nil {
		c.IndentedJSON(http.StatusBadRequest, compileResponse{
			ErrorMessage: err.Error(),
		})
		return
	}

	output, err := puccini.Compile(compileRequest.ServiceTemplateUrl, compileRequest.ScriptletName, compileRequest.ScriptletUrl, compileRequest.Inputs)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, compileResponse{
			Output:       output,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.IndentedJSON(http.StatusOK, compileResponse{
		Output: output,
	})
}

func main() {
	router := gin.Default()
	router.POST("/compile-url", compileFromUrl)

	router.Run("0.0.0.0:8080")
}
