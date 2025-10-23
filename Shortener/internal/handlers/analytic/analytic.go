package analytic

import (
	"fmt"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
)

// Агрегирующая структура для формирования отчет по аналитеке(группировка по текущему дню, прошедший месяц, User-agent)
type AnalyticResponse struct {
	SummaryRedirectCount int
	CurrentDayGroup      []AnalyticUnit
	LastMonthGroup       []AnalyticUnit
	UserAgentGroup       []AnalyticUnit
}

type AnalyticUnit struct {
	Alias       string
	UA          string
	RequestTime time.Time
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
		_, err := storage.GetAnalytic(alias)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
			return
		}
	}
}
