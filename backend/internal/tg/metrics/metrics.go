package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var contactIsOnLine *prometheus.GaugeVec

func Init() {
	contactIsOnLine = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "is_online",
		Help: "One if account is online and zero if it's not",
	}, []string{"account_id", "contact_id", "username"})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:        ":8080",
			ReadTimeout: 3 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
}

func ContactIsOnline(account, contact int64, username string) {
	contactIsOnLine.With(prometheus.Labels{
		"account_id": fmt.Sprintf("%d", account),
		"contact_id": fmt.Sprintf("%d", contact),
		"username":   username,
	}).Set(1)
}

func ContactIsOffline(account, contact int64, username string) {
	contactIsOnLine.With(prometheus.Labels{
		"account_id": fmt.Sprintf("%d", account),
		"contact_id": fmt.Sprintf("%d", contact),
		"username":   username,
	}).Set(0)
}
