package broker

import (
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

type IngestResponse struct {
	Message string `json:"message,omitempty"`
}

func (bg Group) BrokerHandler() gin.HandlerFunc {
	return func(g *gin.Context) {

		brokerID := g.Query("id")

		broker := bg.GetBroker(brokerID)
		if broker == nil {
			err := fmt.Errorf("broker not found")
			g.JSON(400, IngestResponse{
				Message: err.Error(),
			})
			g.Error(err)
			return
		}

		if g.Request.Body == nil {
			err := fmt.Errorf("no payload to process")
			g.JSON(400, IngestResponse{
				Message: err.Error(),
			})
			g.Error(err)
			return
		}

		payload, err := ioutil.ReadAll(g.Request.Body)
		if err != nil {
			g.JSON(400, IngestResponse{
				Message: "unable to read request body",
			})
			g.Error(err)
			return
		}

		if err := broker.Enqueue(payload); err != nil {
			g.Error(err)
			g.JSON(429, IngestResponse{
				Message: "api is overloaded, try again later",
			})
			return
		}

		g.Status(204)
	}
}

const BrokerPath = "/broker"

func (bg Group) RegisterBrokerHandler(engine *gin.Engine) {
	engine.POST(BrokerPath, bg.BrokerHandler())
}
