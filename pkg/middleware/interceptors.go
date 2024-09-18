package middleware

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Shemetov-Sergey/GoNew-service/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	DefaultXRequestIDKey = "x-request-id"
	DefaultXRequestURL   = "x-service-address"
)

func clientInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("Failed at config", err)
	}
	// Logic before invoking the invoker
	meta := metadata.New(map[string]string{})
	meta.Set(DefaultXRequestURL, c.GoNewsSvcUrl+c.Port)
	newCtx := SetRequestId(ctx, meta)
	// Calls the invoker to execute RPC
	err = invoker(newCtx, method, req, reply, cc, opts...)
	return err
}
func WithClientUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(clientInterceptor)
}
func SetRequestId(ctx context.Context, meta metadata.MD) context.Context {
	requestId := HandleRequestID(ctx)
	ctx = metadata.NewOutgoingContext(ctx, meta)
	meta.Set(DefaultXRequestIDKey, requestId)
	return ctx
}
func HandleRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	header, ok := md[DefaultXRequestIDKey]
	if !ok || len(header) == 0 {
		return ""
	}
	requestID := header[0]
	if requestID == "" {
		return ""
	}
	return requestID
}

func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	var requestId string
	requestIdFromContext := metadata.ValueFromIncomingContext(ctx, DefaultXRequestIDKey)

	if len(requestIdFromContext) == 0 {
		requestId = ""
	}

	requestId = metadata.ValueFromIncomingContext(ctx, DefaultXRequestIDKey)[0]

	h, err := handler(ctx, req)

	var address string
	addressFromContext := metadata.ValueFromIncomingContext(ctx, DefaultXRequestURL)
	if len(addressFromContext) == 0 {
		address = ""
	}
	address = metadata.ValueFromIncomingContext(ctx, DefaultXRequestURL)[0]

	//logging
	f, err := os.OpenFile("go-news-svc-request.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	logger := log.New(f, "RequestLogger: ", log.LstdFlags)

	logger.Printf("request - Address:%s\tDuration:%s\trequestId:%s\tError:%v\n",
		address,
		time.Since(start),
		requestId,
		err)

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("Got error %v\n", err)
		}
	}(f)

	return h, err
}
