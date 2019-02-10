package nut

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"strings"
)

type Exporter struct {
	Bind    string
	UPSC    UPSClient
	metrics map[string]Metric
}

type TransformFunc func(string) (float64, error)

type Metric struct {
	lookup    string
	key       string
	multiple  bool
	transform TransformFunc
	gauge     prometheus.Gauge
}

func (e *Exporter) Init() error {
	// https://networkupstools.org/docs/user-manual.chunked/apcs01.html
	keys := []Metric{
		{lookup: "ups.status", key: "ups_online", transform: func(s string) (float64, error) {
			bits := strings.Split(s, " ")
			switch bits[0] {
			case "OL":
				return 1, nil
			case "OB":
				return 0, nil
			case "LB":
				return 0, nil
			default:
				return 0, errors.New("Unknown ups.status state: " + s)
			}
		}},
		{lookup: "ups.status", key: "ups_charging", transform: func(s string) (float64, error) {
			bits := strings.Split(s, " ")
			switch bits[1] {
			// Fully charged?
			case "CHRG":
				return 1, nil
			case "DISCHRG":
				return 0, nil
			default:
				return 0, errors.New("Unknown ups.status state: " + s)
			}
		}},

		{lookup: "battery.capacity", key: "battery_capacity_amp_hours"},
		{lookup: "battery.charge", key: "battery_charge_percent"},
		{lookup: "battery.runtime"},
		{lookup: "input.bypass.frequency"},
		{lookup: "input.bypass.voltage"},
		{lookup: "input.frequency"},
		{lookup: "input.voltage"},

		{lookup: "outlet.current", multiple: true},
		{lookup: "outlet.power", multiple: true},
		{lookup: "outlet.powerfactor", multiple: true},
		{lookup: "outlet.realpower", multiple: true},

		{lookup: "outlet.voltage"},
		{lookup: "outlet.frequency"},
		{lookup: "ups.efficiency"},
		{lookup: "ups.load"},
		{lookup: "ups.realpower"},
		{lookup: "ups.temperature"},
	}

	e.metrics = map[string]Metric{}
	for _, m := range keys {
		if m.key == "" {
			m.key = strings.Replace(m.lookup, ".", "_", -1)
		}

		gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: m.key})
		prometheus.MustRegister(gauge)
		m.gauge = gauge

		e.metrics[m.key] = m
	}

	return nil
}

func (e *Exporter) Poll() error {
	results, err := e.UPSC.All()
	if err != nil {
		return err
	}

	fmt.Println("got results")

	for _, metric := range e.metrics {
		result, ok := results[metric.lookup]
		if !ok {
			fmt.Println("Warning, could not find key in results", metric.lookup)
			continue
		}

		if metric.transform == nil {
			metric.transform = func(s string) (float64, error) {
				return strconv.ParseFloat(s, 64)
			}
		}

		value, err := metric.transform(result)
		if err != nil {
			return err
		}

		metric.gauge.Set(value)
	}

	return nil
}

func (e *Exporter) Listen() error {
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Listening on", e.Bind)
	return http.ListenAndServe(e.Bind, nil)
}
