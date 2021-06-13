package api

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"

	c "github.com/ray-g/dnsproxy/cache"
	"github.com/ray-g/dnsproxy/logger"
	"github.com/ray-g/dnsproxy/stats"
)

// StartAPIServer starts the API server
func StartAPIServer(addr string, debugMode bool, cache c.Cache) error {
	var router *gin.Engine
	if !debugMode {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	} else {
		router = gin.Default()
	}

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	router.Use(cors.Default())

	router.GET("/cache", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"cache": cache.Dump()})
	})

	router.GET("/cache/get/:key", func(c *gin.Context) {
		r, err := cache.Get(c.Param("key"))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": c.Param("key") + " not found"})
		} else {
			c.JSON(http.StatusOK, gin.H{"answer": r.Msg.Answer})
		}
	})

	router.GET("/cache/length", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"length": cache.Length()})
	})

	router.GET("/query/:key", func(c *gin.Context) {
		key := c.Param("key")
		// check cache first
		cr, ce := cache.Get(key)

		// resolve name on localhost
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(key), dns.TypeA)
		clt := new(dns.Client)
		r, _, e := clt.Exchange(m, "127.0.0.1:53")

		resp := gin.H{}
		if e != nil {
			resp["query"] = fmt.Sprintf("failed to resolve %s", key)
		} else if r != nil && r.Rcode != dns.RcodeSuccess {
			resp["query"] = fmt.Sprintf("failed to resolve %s", key)
		} else {
			resp["query"] = r.Answer
		}

		if ce != nil {
			resp["cache"] = fmt.Sprintf("%s not in cache", key)
		} else {
			resp["cache"] = cr.Msg.Answer
		}

		c.JSON(http.StatusOK, resp)
	})

	router.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"stats": stats.Dump()})
	})

	router.GET("/application/active", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"active": stats.Active()})
	})

	router.PUT("/application/active", func(c *gin.Context) {
		active := c.Query("state")
		version := c.Query("v")
		if version != "1" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Illegal value for 'version'"})
		} else {
			switch active {
			case "On":
				stats.Activate()
				c.JSON(http.StatusOK, gin.H{"active": stats.Active()})
			case "Off":
				stats.Deactivate()
				c.JSON(http.StatusOK, gin.H{"active": stats.Active()})
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Illegal value for 'state'"})
			}
		}
	})

	// Serve Web GUI
	router.Use(static.Serve("/", static.LocalFile("./web", false)))

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	logger.Infof("API server listening on %s", addr)
	return err
}
