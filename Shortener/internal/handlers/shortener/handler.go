package shortener

import (
	"fmt"
	"net/http"

	"github.com/v1adis1av28/level3/shortener/internal/utils"
	"github.com/wb-go/wbf/ginext"
)

const ALIAS_STANDART_SIZE = 6

type URLHandler interface {
	GetURL(alias string) (string, error)
	ShortenURL(url, alias string) error
}

type Request struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

// GET
func RedirectShortUrl(c *ginext.Context, storage URLHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		alias := c.Param("short_url")
		fmt.Println(alias)
		if len(alias) < 1 {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "alias url can`t be empty"})
			return
		}
		redireсtUrl, err := storage.GetURL(alias)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, redireсtUrl)
		c.JSON(http.StatusOK, ginext.H{"result": "succesfully redirect"})
	}
}

// POST
func ShortenURL(c *ginext.Context, storage URLHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var req Request
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "encoding json error"})
			return
		}
		if req.Alias == "" { //если пустое название от alias мы должны его сгенериорвать
			req.Alias = utils.GenerateAlias(ALIAS_STANDART_SIZE)
		}

		//todo добавить валидацию url
		err = storage.ShortenURL(req.URL, req.Alias)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "error on shorting url" + err.Error()})
			return
		}

		c.JSON(http.StatusOK, ginext.H{"result": "alias was succesfully register"})
	}
}
