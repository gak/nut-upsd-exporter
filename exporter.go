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
	transform TransformFunc
	gauge     *prometheus.GaugeVec
	labels    []string
}

func (e *Exporter) Init() error {
	// https://networkupstools.org/docs/user-manual.chunked/apcs01.html
	keys := []Metric{
		{lookup: "ups.status", key: "online", transform: func(s string) (float64, error) {
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
		{lookup: "ups.status", key: "charging", transform: func(s string) (float64, error) {
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

		{lookup: "outlet.current", key: "outlet_current", labels: []string{"outlet"}},
		{lookup: "outlet.power", key: "outlet_power", labels: []string{"outlet"}},
		{lookup: "outlet.powerfactor", key: "outlet_powerfactor", labels: []string{"outlet"}},
		{lookup: "outlet.realpower", key: "outlet_realpower", labels: []string{"outlet"}},

		{lookup: "output.voltage"},
		{lookup: "output.frequency"},

		{lookup: "ups.efficiency", key: "efficiency"},
		{lookup: "ups.load", key: "load"},
		{lookup: "ups.realpower", key: "realpower"},
		{lookup: "ups.temperature", key: "temperature"},
	}

	e.metrics = map[string]Metric{}
	for _, m := range keys {
		if m.key == "" {
			m.key = strings.Replace(m.lookup, ".", "_", -1)
		}

		m.key = "ups_" + m.key

		gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: m.key}, m.labels)
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
		value, err := e.value(results, metric, "")
		if err != nil {
			return err
		}

		prefix := "outlet"
		if strings.HasPrefix(metric.lookup, prefix) {
			s := strings.Split(metric.lookup, ".")
			right := s[1]

			for i := 0; i < 3; i++ {
				outlet := ""
				if i > 0 {
					outlet = fmt.Sprintf(".%v", i)
				}
				value, err := e.value(results, metric, prefix+outlet+"."+right)
				if err != nil {
					return err
				}

				metric.gauge.With(prometheus.Labels{"outlet": fmt.Sprintf("%v", i)}).Set(value)
			}
		} else {
			metric.gauge.With(nil).Set(value)
		}
	}

	return nil
}

func (e *Exporter) Listen() error {
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Listening on", e.Bind)
	return http.ListenAndServe(e.Bind, nil)
}

func (e *Exporter) value(results Results, metric Metric, lookupOverride string) (float64, error) {
	if lookupOverride == "" {
		lookupOverride = metric.lookup
	}

	result, ok := results[lookupOverride]
	if !ok {
		fmt.Println(results)
		return 0, errors.New(fmt.Sprintf("could not find key in results %v", lookupOverride))
	}

	if metric.transform == nil {
		metric.transform = func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		}
	}

	return metric.transform(result)
}
