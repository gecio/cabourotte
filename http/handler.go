package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"

	"cabourotte/healthcheck"
)

// BasicResponse a type for HTTP responses
type BasicResponse struct {
	Message string `json:"message"`
}

// addCheck adds a periodic healthcheck to the healthcheck component.
func (c *Component) addCheck(ec echo.Context, healthcheck healthcheck.Healthcheck) error {
	err := c.healthcheck.AddCheck(healthcheck)
	if err != nil {
		msg := fmt.Sprintf("Fail to start the healthcheck: %s", err.Error())
		c.Logger.Error(msg)
		return ec.JSON(http.StatusInternalServerError, &BasicResponse{Message: msg})
	}
	return ec.JSON(http.StatusCreated, &BasicResponse{Message: "Healthcheck successfully added"})
}

// oneOff executes an one-off healthcheck and returns its result
func (c *Component) oneOff(ec echo.Context, healthcheck healthcheck.Healthcheck) error {
	c.Logger.Info(fmt.Sprintf("Executing one-off healthcheck %s", healthcheck.Name()))
	err := healthcheck.Initialize()
	if err != nil {
		msg := fmt.Sprintf("Fail to initialize one off healthcheck %s: %s", healthcheck.Name(), err.Error())
		c.Logger.Error(msg)
		return ec.JSON(http.StatusInternalServerError, &BasicResponse{Message: msg})
	}
	err = healthcheck.Execute()
	if err != nil {
		msg := fmt.Sprintf("Execution of one off healthcheck %s failed: %s", healthcheck.Name(), err.Error())
		c.Logger.Error(msg)
		return ec.JSON(http.StatusInternalServerError, &BasicResponse{Message: msg})
	}
	msg := fmt.Sprintf("One-off healthcheck %s successfully executed", healthcheck.Name())
	c.Logger.Info(msg)
	return ec.JSON(http.StatusCreated, &BasicResponse{Message: msg})
}

// handleCheck handles new healthchecks requests
func (c *Component) handleCheck(ec echo.Context, healthcheck healthcheck.Healthcheck) error {
	if healthcheck.OneOff() {
		return c.oneOff(ec, healthcheck)
	}
	return c.addCheck(ec, healthcheck)
}

// handlers configures the handlers for the http server component
// todo: handler one-off tasks
func (c *Component) handlers() {

	c.Server.POST("/healthcheck/dns", func(ec echo.Context) error {
		var config healthcheck.DNSHealthcheckConfiguration
		if err := ec.Bind(&config); err != nil {
			msg := fmt.Sprintf("Fail to create the dns healthcheck. Invalid JSON: %s", err.Error())
			c.Logger.Error(msg)
			return ec.JSON(http.StatusBadRequest, &BasicResponse{Message: msg})
		}
		healthcheck := healthcheck.NewDNSHealthcheck(c.Logger, &config)
		return c.handleCheck(ec, healthcheck)
	})

	c.Server.POST("/healthcheck/tcp", func(ec echo.Context) error {
		var config healthcheck.TCPHealthcheckConfiguration
		if err := ec.Bind(&config); err != nil {
			msg := fmt.Sprintf("Fail to create the TCP healthcheck. Invalid JSON: %s", err.Error())
			c.Logger.Error(msg)
			return ec.JSON(http.StatusBadRequest, &BasicResponse{Message: msg})
		}
		healthcheck := healthcheck.NewTCPHealthcheck(c.Logger, &config)
		return c.handleCheck(ec, healthcheck)
	})

	c.Server.POST("/healthcheck/http", func(ec echo.Context) error {
		var config healthcheck.HTTPHealthcheckConfiguration
		if err := ec.Bind(&config); err != nil {
			msg := fmt.Sprintf("Fail to create the HTTP healthcheck. Invalid JSON: %s", err.Error())
			c.Logger.Error(msg)
			return ec.JSON(http.StatusBadRequest, &BasicResponse{Message: msg})
		}
		healthcheck := healthcheck.NewHTTPHealthcheck(c.Logger, &config)
		return c.handleCheck(ec, healthcheck)
	})

	c.Server.GET("/healthcheck", func(ec echo.Context) error {
		return ec.JSON(http.StatusOK, c.healthcheck.ListChecks())
	})

	c.Server.DELETE("/healthcheck/:name", func(ec echo.Context) error {
		name := ec.Param("name")
		c.Logger.Info(fmt.Sprintf("Deleting healthcheck %s", name))
		err := c.healthcheck.RemoveCheck(name)
		if err != nil {
			msg := fmt.Sprintf("Fail to start the healthcheck: %s", err.Error())
			c.Logger.Error(msg)
			return ec.JSON(http.StatusInternalServerError, &BasicResponse{Message: msg})
		}
		return ec.JSON(http.StatusOK, &BasicResponse{Message: fmt.Sprintf("Successfully deleted healthcheck %s", name)})
	})
}
