package main

import (
	"bytes"
	"net/http"
)

func appMetricHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var a appMetric

	switch r.Method {
	case "PUT":
		return a.save(r)
	case "GET":
		switch r.Header.Get("Accept") {
		default:
			return a.svg(r, h, b)
		}
	default:
		return &methodNotAllowed
	}
}

func fieldMetricHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var f fieldMetric

	switch r.Method {
	case "PUT":
		return f.save(r)
	case "DELETE":
		return f.delete(r)
	case "GET":
		switch r.Header.Get("Accept") {
		case "text/csv":
			return f.metricCSV(r, h, b)
		default:
			return f.svg(r, h, b)
		}
	default:
		return &methodNotAllowed
	}
}

func fieldTagHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var f fieldTag

	switch r.Method {
	case "PUT":
		return f.save(r, h, b)
	case "DELETE":
		return f.delete(r, h, b)
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			return f.jsonV1(r, h, b)
		default:
			return &notAcceptable
		}

	default:
		return &methodNotAllowed
	}
}

func fieldModelHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var f fieldModel

	switch r.Method {
	case "PUT":
		return f.save(r)
	case "DELETE":
		return f.delete(r)
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			return f.jsonV1(r, h, b)
		default:
			return &notAcceptable
		}
	default:
		return &methodNotAllowed
	}
}

func fieldDeviceHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var f fieldDevice

	switch r.Method {
	case "PUT":
		return f.save(r)
	case "DELETE":
		return f.delete(r)
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			return f.jsonV1(r, h, b)
		default:
			return &notAcceptable
		}
	default:
		return &methodNotAllowed
	}
}

func fieldTypeHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var f fieldType

	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			return f.jsonV1(r, h, b)
		default:
			return &notAcceptable
		}
	default:
		return &methodNotAllowed
	}
}

func fieldThresholdHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var f fieldThreshold

	switch r.Method {
	case "PUT":
		return f.save(r)
	case "DELETE":
		return f.delete(r)
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			return f.jsonV1(r, h, b)
		default:
			return &notAcceptable
		}
	default:
		return &methodNotAllowed
	}
}

func fieldMetricLatestHandler(r *http.Request, h http.Header, b *bytes.Buffer) *result {
	var f fieldLatest

	switch r.Method {
	case "GET":
		switch r.Header.Get("Accept") {
		case "application/json;version=1":
			return f.jsonV1(r, h, b)
		default:
			return f.svg(r, h, b)
		}
	default:
		return &methodNotAllowed
	}
}
