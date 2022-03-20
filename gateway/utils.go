package gateway

import (
	"bytes"
	"encoding/json"
	"github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	logreporter "github.com/openzipkin/zipkin-go/reporter/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func getZipkinTracer(name, addr string, logger *zap.Logger) (*zipkin.Tracer, error) {
	reporter := logreporter.NewReporter(log.New(os.Stderr, "", log.LstdFlags))
	defer reporter.Close()

	// 127.0.0.1:9411
	endpoint, err := zipkin.NewEndpoint(name, addr)

	if err != nil {
		logger.Fatal("无法创建本地终结点: %+v\n", zap.Error(err))
		return nil, err
	}

	// initialize our tracer
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		logger.Fatal("无法创建跟踪器: %+v\n", zap.Error(err))
		return nil, err
	}
	return tracer, nil
}

func NewZipkinServerOption(name, addr string, logger *zap.Logger) (grpc.ServerOption, error) {
	opt, err := getZipkinTracer(name, addr, logger)
	if err != nil {
		return nil, err
	}
	return grpc.StatsHandler(zipkingrpc.NewServerHandler(opt)), nil
}

func NewZipkinClientOption(name, addr string, logger *zap.Logger) (grpc.DialOption, error) {
	opt, err := getZipkinTracer(name, addr, logger)
	if err != nil {
		return nil, err
	}
	return grpc.WithStatsHandler(zipkingrpc.NewClientHandler(opt)), nil
}

func ReadDataFromBody(w http.ResponseWriter, req *http.Request) (map[string]interface{}, error) {
	var mapResult map[string]interface{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &mapResult)
	if err != nil {
		return nil, err
	}
	// 还原Body
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return mapResult, nil
}
