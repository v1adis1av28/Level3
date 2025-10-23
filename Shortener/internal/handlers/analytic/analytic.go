package analytic

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
)

// Агрегирующая структура для формирования отчет по аналитеке(группировка по текущему дню, прошедший месяц, User-agent)
type AnalyticResponse struct {
	SummaryRedirectCount int              `json:"redirect-count"`
	CurrentDayGroup      []AnalyticUnit   `json:"current-day"`
	LastMonthGroup       []AnalyticUnit   `json:"last-month"`
	UserAgentGroup       []UAAnaliticUnit `json:"user-agent"`
}

type AnalyticUnit struct {
	Alias       string    `json:"alias"`
	UA          string    `json:"user-agent"`
	RequestTime time.Time `json:"request-time"`
}

type UAAnaliticUnit struct {
	Alias         string      `json:"alias"`
	UA            string      `json:"user-agent"`
	RequestTimes  []time.Time `json:"request-time"`
	RedirectCount int         `json:"redirect-count"`
}

type AnalyticHandler interface {
	GetAnalytic(alias string) (*AnalyticResponse, error)
}

func GetAnalytic(c *ginext.Context, storage AnalyticHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		alias := c.Param("short_url")
		fmt.Println(alias)
		if len(alias) < 1 {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "alias url can`t be empty"})
			return
		}
		analytic, err := storage.GetAnalytic(alias)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ginext.H{"result": analytic})
		return
	}
}
